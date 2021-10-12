#!/bin/bash

function clean {
    local PID=$1
    kill -9 ${PID}
    echo_time "Searching for running noobaa and killing it."
    local process=${noobaa// */}
    local kill_noobaa_cli_pid=($(ps -ef | grep ${process} | grep -v grep | awk '{print $2}'| xargs))
    if [ -z ${kill_noobaa_cli_pid} ]
    then
        kill -9 ${kill_noobaa_cli_pid[@]}
    fi
    exit 0
}

function kuberun {
    if [ "${1}" == "silence" ]
    then
        silence=true
        shift
    fi
    local options=$*
    if [ -z "${kubectl}" ]
    then
        echo_time "❌  The kubectl variable must be define in the shell"
        exit 1
    fi
    ${kubectl} ${options}
    if [ $? -ne 0 ]
    then
        echo_time "❌  ${kubectl} ${options} failed, Exiting"
        exit 1
    elif [ ! ${silence} ]
    then
        echo_time "✅  ${kubectl} ${options} passed"
    fi
}

echo_time() {
    date +"%T $*"
}

function test_noobaa {
    local rc func timeout_in_sec
    local {timeout,should_fail,silence}=false
    local count=1
    local retries=18

    if [[ "${1}" =~ ("should_fail"|"silence") ]]
    then
        eval ${1}=true
        shift
    fi

    while true
    do
        if [[ ! "${1}" =~ "--" ]]
        then
            break
        fi

        case ${1} in
            --func)         func="--func ${2}"
                            shift 2;;
            --timeout)      timeout=true
                            if [[ "${2}" =~ "--" ]] || [[ ! "${2}" =~ ^[0-9]+$ ]]
                            then
                                shift 1
                            else
                                timeout_in_sec="--timeout ${2}"
                                shift 2
                            fi;;
            *)              echo_time "❌  Unknown test_noobaa option, Exiting."
                            exit 1;;
        esac
    done

    local options=$*

    if ${timeout}
    then
        ${noobaa} ${options} &
        local PID=$!
        # We are trapping SIGHUP and SIGINT for clean exit.
        trap "clean ${PID}" 1 2
        # When we are running with timeout because the command runs in the background
        timeout --PID ${PID} ${timeout_in_sec} ${func} ${options}
    else
        local rc=1
        while [ $rc -ne 0 ]
        do
            ${noobaa} ${options}
            rc=$?
            if [ $rc -ne 0 ]
            then
                if ${should_fail}
                then
                    echo_time "✅  ${noobaa} ${options} failed - as should"
                    rc=0
                else 
                    if [ ${count} -lt ${retries} ]
                    then
                        echo_time "❌ failed to run ${noobaa} ${options} retrying" 
                        sleep 10
                        count=$((count+1))
                    else
                        echo_time "❌  ${noobaa} ${options} failed, Exiting"
                        local pod_operator=$(kuberun get pod | grep noobaa-operator | awk '{print $1}')
                        echo_time "==============OPERATOR LOGS============"
                        kuberun logs ${pod_operator}
                        echo_time "==============CORE LOGS============"
                        kuberun logs noobaa-core-0
                        exit 1
                    fi
                fi
            elif [ ! ${silence} ]
            then
                echo_time "✅  ${noobaa} ${options} passed"
            fi
        done
    fi

}

function timeout {
    local PID func
    #the timeout is that big because it sometimes take a while to get pvc
    local TIMEOUT=180
    while true
    do
        if [[ ! "${1}" =~ "--" ]]
        then
            break
        fi

        case ${1} in
            --PID)      PID=${2}
                        shift 2;;
            --func)     func=${2}
                        shift 2;;
            --timeout)  TIMEOUT=${2}
                        shift 2;;
            *)          echo_time "❌  Unknown timeout option, Exiting."
                        exit 1;;
        esac
    done
    local options=$*
    local START_TIME=${SECONDS}

    while true
    do
        kill -s 0 ${PID} &> /dev/null
        if [ $? -ne 0 ]
        then
            echo_time "✅  ${noobaa} ${options} passed"
            break
        fi

        if [ $((START_TIME+TIMEOUT)) -gt ${SECONDS} ]
        then
            sleep 5
        else
            kill -9 ${PID}
            if [ ! -z ${func} ]
            then
                echo_time "${noobaa} ${options} reached timeout, Running ${func}"
                ${func}
            fi 
            echo_time "❌  ${noobaa} ${options} reached timeout, Exiting"
            exit 1
        fi
    done
}

function install {
    local use_obc_cleanup_policy
    
    [ $((RANDOM%2)) -gt 0 ] && use_obc_cleanup_policy="--use-obc-cleanup-policy"
    test_noobaa install --mini ${use_obc_cleanup_policy}

    local status=$(kuberun silence get noobaa noobaa -o 'jsonpath={.status.phase}')
    while [ "${status}" != "Ready" ]
    do
        echo_time "💬  Waiting for status Ready, Status is ${status}"
        sleep 10
        status=$(kuberun silence get noobaa noobaa -o 'jsonpath={.status.phase}')
    done
}

function noobaa_install {
    #noobaa timeout install # Maybe when creating server we can use local PV
    install
    test_noobaa status
    kuberun get noobaa
    kuberun describe noobaa
}

function check_core_config_map {
    kuberun get configmap noobaa-config
    check_change_debug_level_in_config_map
}

function check_change_debug_level_in_config_map {
    local cm_debug_level="all"
    local patch='{"data":{"NOOBAA_LOG_LEVEL":"all"}}'
    local timeout=0
    local core_debug_level=$(kuberun silence exec noobaa-core-0 -- printenv NOOBAA_LOG_LEVEL)

    kuberun silence patch configmap noobaa-config -p ${patch}

    while [[ "${core_debug_level}" != "${cm_debug_level}" ]]
    do
        echo_time "💬  Waiting for NOOBAA_LOG_LEVEL core env var to match the noobaa-config"
        timeout=$((timeout+10))
        sleep 10
        core_debug_level=$(kuberun silence exec noobaa-core-0 -- printenv NOOBAA_LOG_LEVEL)
        if [ ${timeout} -ge 180 ] 
        then
            echo_time "❌  reached the timeout for waiting to the update"
            break
        fi
    done 

    if [[ "${core_debug_level}" == "${cm_debug_level}" ]]
    then
        echo_time "✅  noobaa core env variable updated successfully"
    else
        echo_time "❌  noobaa core env var NOOBAA_LOG_LEVEL didn't got updated, Exiting"
        exit 1
    fi
}

function aws_credentials {
    while read line
    do
        if [[ ${line} =~ (AWS_ACCESS_KEY_ID|AWS_SECRET_ACCESS_KEY) ]]
        then
            eval $(echo ${line//\"/} | sed -e 's/ //g' -e 's/:/=/g')
        fi
    done < <(test_noobaa silence status)
    if [ -z ${AWS_ACCESS_KEY_ID} ] || [ -z ${AWS_SECRET_ACCESS_KEY} ]
    then
        echo_time "❌  Could not get AWS credentials, Exiting"
        exit 1
    fi
}

function check_namespacestore {
    echo_time "💬  Staring namespacestore cycle"
    local cycle
    local type="s3-compatible"
    local buckets=("target.bucket1" "target.bucket2")
    local namespacestore=("namespacestore5" "namespacestore6")

    test_noobaa bucket create ${buckets[0]}
    test_noobaa bucket create ${buckets[1]}

    for (( cycle=0 ; cycle < ${#namespacestore[@]} ; cycle++ ))
    do
        test_noobaa namespacestore create ${type} ${namespacestore[cycle]} \
            --target-bucket ${buckets[cycle]} \
            --endpoint s3.${NAMESPACE}.svc.cluster.local:443 \
            --access-key ${AWS_ACCESS_KEY_ID} \
            --secret-key ${AWS_SECRET_ACCESS_KEY}
        test_noobaa namespacestore status ${namespacestore[cycle]}
    done
    
    test_noobaa namespacestore list
    test_noobaa status
    kuberun get namespacestore
    kuberun describe namespacestore

    check_namespacestore_validator

    echo_time "✅  namespace store s3 compatible cycle is done"
}

function check_namespacestore_validator {
    check_namespacestore_nsfs_validator
}

function check_namespacestore_nsfs_validator {
    echo_time "💬  Staring namespacestore nsfs validator cycle"

    #Setup
    local type="nsfs"
    local pvc="nsfs-vol"
    local namespacestore="namespacestore-"${type}

    kuberun create -f $(dirname ${0})/resources/nsfs-local-class.yaml
    kuberun create -f $(dirname ${0})/resources/nsfs-local-pv.yaml
    kuberun create -f $(dirname ${0})/resources/nsfs-local-pvc.yaml
    
    #Sub-path is not relative
    test_noobaa should_fail namespacestore create ${type} ${namespacestore} \
        --fs-backend 'GPFS' \
        --pvc-name ${pvc} \
        --sub-path '/'
    
    #Sub-path contains '..'
    test_noobaa should_fail namespacestore create ${type} ${namespacestore} \
        --fs-backend 'GPFS' \
        --pvc-name ${pvc} \
        --sub-path 'subpath/../'

    #Valid sub-path
    test_noobaa namespacestore create ${type} ${namespacestore} \
        --fs-backend 'GPFS' \
        --pvc-name ${pvc} \
        --sub-path 'subpath'
    
    test_noobaa namespacestore list

    #cleanup
    test_noobaa silence namespacestore delete ${namespacestore}


    echo_time "✅  namespacestore nsfs validator is done"
}

function check_S3_compatible {
    echo_time "💬  Staring compatible cycle"
    local cycle
    local type="s3-compatible"
    local buckets=("first.bucket" "second.bucket")
    local backingstore=("compatible1" "compatible2")

    test_noobaa bucket create ${buckets[1]}
    test_noobaa backingstore create pv-pool pvpool1 \
            --num-volumes 1 \
            --pv-size-gb 50

    for (( cycle=0 ; cycle < ${#backingstore[@]} ; cycle++ ))
    do
        test_noobaa backingstore create ${type} ${backingstore[cycle]} \
            --target-bucket ${buckets[cycle]} \
            --endpoint s3.${NAMESPACE}.svc.cluster.local:443 \
            --access-key ${AWS_ACCESS_KEY_ID} \
            --secret-key ${AWS_SECRET_ACCESS_KEY}
        test_noobaa backingstore status ${backingstore[cycle]}
    done
    test_noobaa backingstore list
    test_noobaa status
    kuberun get backingstore
    kuberun describe backingstore
    echo_time "✅  s3 compatible cycle is done"
}

function check_IBM_cos {
    echo_time "💬  Staring IBM cos cycle"
    local cycle
    local type="ibm-cos"
    local buckets=("first.bucket" "second.bucket")
    local backingstore=("ibmcos1" "ibmcos2")

    test_noobaa bucket create ${buckets[1]}
    for (( cycle=0 ; cycle < ${#backingstore[@]} ; cycle++ ))
    do
        test_noobaa backingstore create ${type} ${backingstore[cycle]} \
            --target-bucket ${buckets[cycle]} \
            --endpoint s3.${NAMESPACE}.svc.cluster.local:443 \
            --access-key ${AWS_ACCESS_KEY_ID} \
            --secret-key ${AWS_SECRET_ACCESS_KEY}
        test_noobaa backingstore status ${backingstore[cycle]}
    done
    test_noobaa backingstore list
    test_noobaa status
    kuberun get backingstore
    kuberun describe backingstore
    echo_time "✅  ibm cos cycle is done"
}

function check_aws_S3 {
    return
    # test_noobaa bucket create second.bucket
    # test_noobaa backingstore create aws1 --type aws-s3 --bucket-name znoobaa --access-key XXX --secret-key YYY
    # test_noobaa backingstore create aws2 --type aws-s3 --bucket-name noobaa-qa --access-key XXX --secret-key YYY
    # test_noobaa backingstore status aws1
    # test_noobaa backingstore status aws2
    # test_noobaa backingstore list
    # test_noobaa status
    # kubectl get backingstore
    # kubectl describe backingstore
}

function bucketclass_cycle {
    echo_time "💬  Starting the bucketclass cycle"
    local bucketclass
    local bucketclass_names=()
    local backingstore=()
    local namespacestore=()
    local number_of_backingstores=4
    local number_of_namespacestores=2

    for (( number=0 ; number <= (number_of_backingstores + number_of_namespacestores); number++ ))
    do
        bucketclass_names+=("bucket.class$((number+1))")
        if [ "$number" -lt "$number_of_backingstores" ]
        then
            backingstore+=("compatible$((number+1))")
        else
            namespacestore+=("namespacestore$((number+1))")
        fi
    done
    


    test_noobaa bucketclass create placement-bucketclass ${bucketclass_names[0]} --backingstores ${backingstore[0]}
    # test_noobaa bucketclass create placement-bucketclass ${bucketclass_names[1]} --placement Mirror --backingstores nb1,aws1 ❌
    # test_noobaa bucketclass create placement-bucketclass ${bucketclass_names[2]} --placement Spread --backingstores aws1,aws2 ❌
    test_noobaa bucketclass create placement-bucketclass ${bucketclass_names[3]} --backingstores ${backingstore[0]},${backingstore[1]}   
    test_noobaa bucketclass create namespace-bucketclass single ${bucketclass_names[4]} --resource ${namespacestore[0]}
    test_noobaa bucketclass create namespace-bucketclass multi ${bucketclass_names[5]} --read-resources ${namespacestore[0]},${namespacestore[1]} --write-resource ${namespacestore[0]} 
    test_noobaa bucketclass create namespace-bucketclass cache ${bucketclass_names[6]} --hub-resource ${namespacestore[1]} --backingstores ${backingstore[1]}
    test_noobaa bucketclass create placement-bucketclass "bucket.class.replication" --backingstores ${backingstore[0]} --replication-policy replication1.json
    bucketclass_names+=("bucket.class.replication")

    local bucketclass_list_array=($(test_noobaa silence bucketclass list | awk '{print $1}' | grep -v NAME))
    for bucketclass in ${bucketclass_list_array[@]}
    do
        test_noobaa bucketclass status ${bucketclass}
    done

    #TODO: activate the code below when we create all the bucketclass
    # if [ ${#bucketclass_list_array[@]} -ne $((${#bucketclass_names[@]}+1)) ]
    # then
    #     echo_time "❌  Bucket expected $((${#bucketclass_names[@]}+1)), and got ${#bucketclass_list_array[@]}."
    #     echo_time "👓  bucketclass list is ${bucketclass_list_array[@]}, Exiting."
    #     exit 1
    # fi

    test_noobaa status
    kuberun get bucketclass
    kuberun describe bucketclass
    echo_time "✅  bucketclass cycle is done"
}

function check_obc {
    local bucket
    test_noobaa obc list
    for bucket in ${buckets[@]}
    do
        test_noobaa --timeout obc status ${bucket}
    done
    kuberun get obc
    kuberun describe obc
    kuberun get obc,ob,secret,cm -l noobaa-obc
}

function obc_cycle {
    echo_time "💬  Starting the obc cycle"
    local buckets=()

    local bucketclass_list_array=($(test_noobaa silence bucketclass list | awk '{print $1}' | grep -v NAME | grep -v noobaa-default-bucket-class))
    for bucketclass in ${bucketclass_list_array[@]}
    do
        buckets+=("bucket${bucketclass//[a-zA-Z.-]/}")
        if [ "${bucketclass//[a-zA-Z.-]/}" == "3" ]
        then
            flag="--app-namespace default"
        fi
        # for bucketclass7 - create 2 obcs, one using its one replication policy and second one that using the bucketclass replication and 
        if [ "${bucketclass//[a-zA-Z.-]/}" == "7" ]
        then
            flag="--replication-policy replication_policy2.json"
            test_noobaa --timeout --func check_obc obc create "${buckets[$((${#buckets[@]}-1))]}_obc_repl" --bucketclass ${bucketclass} ${flag}
            unset flag
        fi
        test_noobaa --timeout --func check_obc obc create ${buckets[$((${#buckets[@]}-1))]} --bucketclass ${bucketclass} ${flag}
        unset flag
    done
    check_obc

    # aws s3 --endpoint-url XXX ls
    echo_time "✅  obc cycle is done"
}

function delete_backingstore_path {
    local object_bucket backing_store
    local backingstore=($(test_noobaa silence backingstore list | grep -v "NAME" | awk '{print $1}'))
    local bucketclass=($(test_noobaa silence bucketclass list  | grep ${backingstore[1]} | awk '{print $1}'))
    local obc=()
    local all_obc=($(test_noobaa silence obc list | grep -v "BUCKET-NAME" | awk '{print $2":"$5}'))
    
    # get obcs that their bucketclass is in bucketclass array
    for object_bucket in ${all_obc[@]}
    do
        local cur_bucketclass=($(awk -F: '{print $2}' <<< ${object_bucket}))
        local cur_obc_name=($(awk -F: '{print $1}' <<< ${object_bucket}))
        for bucket_class in ${bucketclass[@]}
        do
            if [[ ${cur_bucketclass} == ${bucket_class} ]]
            then
                obc+=(${cur_obc_name})
            fi
        done
    done

    test_noobaa should_fail backingstore delete ${backingstore[1]}
    if [ ${#obc[@]} -ne 0 ]
    then
        for object_bucket in ${obc[@]}
        do
            test_noobaa obc delete ${object_bucket}
        done
    fi
    if [ ${#bucketclass[@]} -ne 0 ]
    then
        for bucket_class in ${bucketclass[@]}
        do
            test_noobaa bucketclass delete ${bucket_class}
        done
    fi
    sleep 30
    local buckets=($(test_noobaa silence bucket list  | grep -v "BUCKET-NAME" | awk '{print $1}'))
    echo_time "✅  buckets in system: ${buckets}"
    test_noobaa backingstore delete ${backingstore[1]}
    test_noobaa should_fail backingstore delete ${backingstore[0]}
    echo_time "✅  delete ${backingstore[1]} path is done"
}

function delete_namespacestore_path {
    local object_bucket namespace_store
    test_noobaa obc delete ${obc[2]}
    test_noobaa bucketclass delete ${bucketclass[2]}
    local namespacestore=($(test_noobaa silence namespacestore list | grep -v "NAME" | awk '{print $1}'))
    local bucketclass=($(test_noobaa silence bucketclass list | grep -v "NAME" | awk '{print $1}'))
    local obc=()
    local all_obc=($(test_noobaa silence obc list | grep -v "BUCKET-NAME" | awk '{print $2":"$5}'))
    
    # get obcs that their bucketclass is in bucketclass array
    for object_bucket in ${all_obc[@]}
    do
        local cur_bucketclass=($(awk -F: '{print $2}' <<< ${object_bucket}))
        local cur_obc_name=($(awk -F: '{print $1}' <<< ${object_bucket}))
        for bucket_class in ${bucketclass[@]}
        do
            if [[ ${cur_bucketclass} == ${bucket_class} ]]
            then
                obc+=(${cur_obc_name})
            fi
        done
    done

    echo_time "💬  Starting the delete related ${namespacestore[1]} paths"

    test_noobaa should_fail namespacestore delete ${namespacestore[1]}
    if [ ${#obc[@]} -ne 0 ]
    then
        for object_bucket in ${obc[@]}
        do
            test_noobaa obc delete ${object_bucket}
        done
    fi
    if [ ${#bucketclass[@]} -ne 0 ]
    then
        for bucket_class in ${bucketclass[@]}
        do
            test_noobaa bucketclass delete ${bucket_class}
        done
    fi
    sleep 30
    local buckets=($(test_noobaa silence bucket list  | grep -v "BUCKET-NAME" | awk '{print $1}'))
    echo_time "✅  buckets in system: ${buckets}"
    test_noobaa namespacestore delete ${namespacestore[0]}
    test_noobaa namespacestore delete ${namespacestore[1]}
    echo_time "✅  delete ${namespacestore[1]} and ${namespacestore[0]} path is done"
}

function check_deletes {
    echo_time "💬  Starting the delete cycle"
    local obc=($(test_noobaa silence obc list | grep -v "NAME\|default" | awk '{print $2}'))
    local bucketclass=($(test_noobaa silence bucketclass list  | grep -v NAME | awk '{print $1}'))
    local backingstore=($(test_noobaa silence backingstore list | grep -v "NAME" | awk '{print $1}'))
    test_noobaa obc delete ${obc[0]}
    test_noobaa bucketclass delete ${bucketclass[0]}
    test_noobaa backingstore list
    delete_backingstore_path
    delete_namespacestore_path
    echo_time "✅  delete cycle is done"
}

function crd_arr { 
    crd_array=($(kubectl get crd | awk '{print $1}' | grep -v "NAME"))
    echo_time "${crd_array[*]}"
}
function noobaa_uninstall {
    local cleanup cleanup_data
    local check_cleanflag=$((RANDOM%2))
    local check_cleanup_data_flag=$((RANDOM%2))

    [ ${check_cleanflag} -eq 0 ] &&  cleanup="--cleanup"
    [ ${check_cleanup_data_flag} -eq 0 ] && cleanup_data="--cleanup_data"

    echo_time "💬  Running uninstall ${cleanup} ${cleanup_data}"
    yes | test_noobaa --timeout uninstall ${cleanup} ${cleanup_data}
    if [ ${check_cleanflag} -eq 0 ]
    then
        check_if_cleanup
    fi
}

function check_if_cleanup {  
    crd_array_after_Cleanup=($(kubectl get crd | awk '{print $1}' | grep -v "NAME"))   
    for crd_before_clean in ${crd_array[@]}
    do
        if [[ ${crd_array_after_Cleanup[@]} =~ ${crd_before_clean} ]]
        then
            echo_time "${crd_before_clean} is in crd"
            exit 1   
        else         
            echo_time "${crd_before_clean} is not in crd, deleted with clenaup"
        fi               
    done

    for name in ${crd_array[@]} 
    do
        noobaa crd status &>/dev/stdout | grep -v "Not Found" | grep -q "${name}"
        if [ $? -ne 0 ]  
        then    
            echo_time "${name} crd status empty"     
        else 
            echo_time "${name} crd status not empty" 
            exit 1    
        fi
    done
    
    kubectl get namespace ${NAMESPACE}
    if [ $? -ne 0 ] 
    then   
        echo_time "namespace doesnt exist" 
    else
        echo_time "namespace still exists"
        exit 1            
    fi
} 

if [ -z "${noobaa}" ]
then
    echo_time "❌  The noobaa variable must be define in the shell"
    exit 1
fi


function create_replication_files {
    echo "[{ \"rule_id\": \"rule-1\", \"destination_bucket\": \"first.bucket\", \"filter\": {\"prefix\": \"d\"}} ]" > replication1.json
    echo "[{ \"rule_id\": \"rule-2\", \"destination_bucket\": \"first.bucket\", \"filter\": {\"prefix\": \"e\"}} ]" > replication2.json
}

function delete_replication_files {
    rm "replication1.json"
    rm "replication2.json"
}

function check_backingstore {
    echo_time "💬  Creating bucket testbucket"
    test_noobaa bucket create "testbucket"

    local tier=`noobaa api bucket_api read_bucket '{ "name": "testbucket" }' | grep -w "tier" | awk '{ print $2 }'`
    local bs=`noobaa api tier_api read_tier '{ "name": "'$tier'" }' | grep -m 1 "noobaa-default-backing-store"` 

    if [ ! -z "$bs" ]
    then
        echo_time "❌  backingstore for the bucket is not the default backingstore"
        exit 1
    fi

    echo_time "💬  Deleting bucket testbucket"
    test_noobaa bucket delete "testbucket"
}

function check_dbdump {
    echo_time "💬  Generating db dump through dump command"

    # Generate db dump at /tmp
    test_noobaa db-dump --dir /tmp

    # Check whether dump was created
    local dump_file_name=`ls -l /tmp | grep noobaa_db_dump | awk '{ print $9 }'`
    if [ ! -f "/tmp/$dump_file_name" ]
    then
        echo_time "❌  db dump was not generated through dump command"
        exit 1
    fi

    # Remove dump file
    rm /tmp/$dump_file_name

    # Generate db dump through diagnose API
    echo_time "💬  Generating db dump through diagnose command"
    test_noobaa diagnose --db-dump --dir /tmp

    # Check whether dump was created
    local diagnose_file_name=`ls -l /tmp | grep noobaa_diagnostics | awk '{ print $9 }'`
    dump_file_name=`ls -l /tmp | grep noobaa_db_dump | awk '{ print $9 }'`
    if [ ! -f "/tmp/$dump_file_name" ]
    then
        echo_time "❌  db dump was not generated through diagnose command"
        exit 1
    fi

    # Remove diagnostics and dump files
    rm /tmp/$diagnose_file_name
    rm /tmp/$dump_file_name
}
