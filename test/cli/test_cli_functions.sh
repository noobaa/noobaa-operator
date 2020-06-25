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
        ${noobaa} ${options}
        if [ $? -ne 0 ]
        then
            if ${should_fail}
            then
                echo_time "✅  ${noobaa} ${options} failed - as should"
            else 
                echo_time "❌  ${noobaa} ${options} failed, Exiting"
                local pod_operator=$(kuberun get pod | grep noobaa-operator | awk '{print $1}')
                echo_time "==============OPERATOR LOGS============"
                kuberun logs ${pod_operator}
                echo_time "==============CORE LOGS============"
                kuberun logs noobaa-core-0
                exit 1
            fi
        elif [ ! ${silence} ]
        then
            echo_time "✅  ${noobaa} ${options} passed"
        fi
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
    local number_of_backingstores=4

    for (( number=0 ; number < number_of_backingstores ; number++ ))
    do
        bucketclass_names+=("bucket.class$((number+1))")
        backingstore+=("compatible$((number+1))")
    done

    test_noobaa bucketclass create ${bucketclass_names[0]} --backingstores ${backingstore[0]}
    # test_noobaa bucketclass create ${bucketclass_names[1]} --placement Mirror --backingstores nb1,aws1 ❌
    # test_noobaa bucketclass create ${bucketclass_names[2]} --placement Spread --backingstores aws1,aws2 ❌
    test_noobaa bucketclass create ${bucketclass_names[3]} --backingstores ${backingstore[0]},${backingstore[1]}

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
        test_noobaa --timeout --func check_obc obc create ${buckets[$((${#buckets[@]}-1))]} --bucketclass ${bucketclass} ${flag}
        unset flag
    done
    check_obc

    # aws s3 --endpoint-url XXX ls
    echo_time "✅  obc cycle is done"
}

function crd_cycle {     
    echo_time "Starting the crd cycle" 
    local crd_create_array=($(test_noobaa silence crd create &>/dev/stdout | awk '{print $7}' | sed -e 's/[""\\]//g' | sed 's/.$//'))
    local crd_status_array=($(test_noobaa silence crd status &>/dev/stdout | awk '{print $6}' | grep -v "Exists" | sed -e 's/[""\\]//g' | sed 's/.$//'))
    crd_array=($(kubectl get crd | awk '{print $1}' | grep -v "NAME" | grep -e "noobaa" -e "objectbucket"))

    #comparing crd status and create arrays
    echo_time "checking if crds match in two noobaa crd commands"
    for crd_status_array in ${crd_array[@]}
    do
        if [[ ${crd_create_array[@]} =~ ${crd_status_array} ]]
        then            
            echo_time "✅ ${crd_status_array} exists in the crds list"
        else
            echo_time "❌ ${crd_status_array} is not existed in the crds list"  
            exit 1       
        fi               
    done  

    #TODO:find a way to ac tivate CRD delete flow after noobaa install 
    #test_noobaa timeout crd delete 
    #local crd_status_after_delete=($(test_noobaa silence crd status &>/dev/stdout | awk '{print $6}' | sed -e 's/[""\\]//g' | sed 's/.$//'))
    #checking if crds still exist in the test after the delete
    #for crd in ${crd_array[@]} 
    #do
    #    if [[ ${crd_status_after_delete[@]} =~ ${crd_status_array} ]]
    #    then    
    #        echo_time "❌ crd ${crd} still exists in the test. exiting."
    #        exit 1       
    #    else 
    #        echo_time "✅ crd ${crd} deleted from test" 
    #    fi
    #done

    echo_time "creating crd again for checks..."     
    test_noobaa timeout crd create 
    local crd_after_create=($(test_noobaa silence crd status &>/dev/stdout | awk '{print $6}' | sed -e 's/[""\\]//g' | sed 's/.$//'))
    #comparing crds after running create command
    for crd_after_create in ${crd_array[@]}
    do
        if [[ ${crd_after_create[@]} =~ ${crd_status_array} ]]
        then            
            echo_time "✅ ${crd_after_create} is exsisted in the crds list"
        else
            echo_time "❌ ${crd_after_create} is not existed in the crds list"
            exit 1         
        fi               
    done

    echo_time "✅  crd cycle is done"
}

function delete_backingstore_path {
    local object_bucket backing_store
    local backingstore=($(test_noobaa silence backingstore list | grep -v "NAME" | awk '{print $1}'))
    local bucketclass=($(test_noobaa silence bucketclass list  | grep ${backingstore[1]} | awk '{print $1}'))
    local obc=($(test_noobaa silence obc list | grep -v "BUCKET-NAME" | awk '{print $2}'))
    echo_time "💬  Starting the delete related ${backingstore[1]} paths"

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

function check_deletes {
    echo_time "💬  Starting the delete cycle"
    local obc=($(test_noobaa silence obc list | grep -v "NAME\|default" | awk '{print $2}'))
    local bucketclass=($(test_noobaa silence bucketclass list  | grep -v NAME | awk '{print $1}'))
    local backingstore=($(test_noobaa silence backingstore list | grep -v "NAME" | awk '{print $1}'))
    test_noobaa obc delete ${obc[0]}
    test_noobaa bucketclass delete ${bucketclass[0]}
    test_noobaa backingstore list
    delete_backingstore_path
    echo_time "✅  delete cycle is done"
}

function noobaa_uninstall {
    check_cleanflag=$(( RANDOM%2 ))
    echo_time $check_cleanflag
    if [ ${check_cleanflag} -gt 0 ] 
    then
        echo_time "Running uninstall with --cleanup”"
        test_noobaa uninstall --cleanup 
    else
        echo_time "Running uninstall without --cleanup”"
        test_noobaa uninstall 
    fi  
}
function check_if_cleanup {  
    crd_array_after_Cleanup=($(kubectl get crd | awk '{print $1}' | grep -v "NAME"))
  for empty_crd in ${crd_array[@]}; do
    in=false
        for crd_afrer in "${crd_array_after_Cleanup[@]}"; do
            if [[ $empty_crd == $crd_after ]]; then
                echo "$empty_crd is in crd"
                in=true
                break
            fi
        done
    $in || echo "$empty_crd is not in crd, deleted with clenaup"
    done
    for name in ${crd_array[@]} 
    do
        noobaa crd status &>/dev/stdout | grep -v "Not Found" | grep -q "$name"
        if [ $? -ne 0 ]  
         then    
            echo "$name crd status empty"     
        else 
            echo "$name crd status not empty"    
        fi
    done
    kubectl get namespace ${NAMESPACE}
    if [ $? -ne 0 ] 
    then   
        echo "namespace doesnt exist" 
    else
        echo "namespace still exists"           
    fi    
} 