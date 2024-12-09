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
                        kuberun logs ${pod_operator} -c noobaa-operator
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
    test_noobaa install --${RESOURCE} --admission ${use_obc_cleanup_policy}

    wait_for_noobaa_ready
    wait_for_backingstore_ready noobaa-default-backing-store
}

function run_external_postgres {
    # kubectl run postgres-external --image=postgres:15 --env POSTGRES_PASSWORD=password --port 5432 --expose
    echo_time "Creating an external postgres DB for test (NO SSL)"
    kuberun create -f $(dirname ${0})/resources/external-db.yaml
}

function run_external_postgres_ssl {
    echo_time "Creating an external postgres DB for test (SSL)"
    kuberun create secret generic postgres-ssl --from-file=certs/server.crt --from-file=certs/server.key --from-file=certs/ca.crt
    kuberun create -f $(dirname ${0})/resources/external-db-ssl.yaml
}

function delete_external_postgres {
    kuberun delete -f $(dirname ${0})/resources/external-db.yaml
}

function delete_external_postgres_ssl {
    kuberun delete -f $(dirname ${0})/resources/external-db-ssl.yaml
    kuberun delete secret postgres-ssl
}

function install_external {    
    local postgres_url="postgresql://postgres:password@postgres-external.${NAMESPACE}.svc:5432/postgres"
    echo_time "Installing NooBaa in external postgres mode postgres-url=${postgres_url}"
    test_noobaa install --${RESOURCE} --postgres-url=${postgres_url}

    wait_for_noobaa_ready
    wait_for_backingstore_ready noobaa-default-backing-store
}

function install_external_ssl {    
    local postgres_url="postgresql://postgres:password@postgres-external.${NAMESPACE}.svc:5432/postgres"
    echo_time "Installing NooBaa in external postgres mode postgres-url=${postgres_url} with SSL"
    test_noobaa install --${RESOURCE} --postgres-url=${postgres_url} --pg-ssl-required --pg-ssl-unauthorized --pg-ssl-key certs/client.key --pg-ssl-cert certs/client.crt

    wait_for_noobaa_ready
    wait_for_backingstore_ready noobaa-default-backing-store
}

function wait_for_noobaa_ready {
    local status=$(kuberun silence get noobaa noobaa -o 'jsonpath={.status.phase}')
    while [ "${status}" != "Ready" ]
    do
        echo_time "💬  Waiting for status Ready, Status is ${status}"
        sleep 10
        status=$(kuberun silence get noobaa noobaa -o 'jsonpath={.status.phase}')
    done
}

function wait_for_backingstore_ready {
    local status=$(kuberun silence get backingstore noobaa-default-backing-store -o 'jsonpath={.status.phase}')
    local status=$(kuberun silence get backingstore ${1} -o 'jsonpath={.status.phase}')
    while [ "${status}" != "Ready" ]
    do
        echo_time "💬  Waiting for status Ready, Status is ${status}"
        sleep 10
        status=$(kuberun silence get noobaa noobaa -o 'jsonpath={.status.phase}')
    done
}

function clean_leftovers {
    test_noobaa --timeout uninstall
    kuberun delete deploy,sts,service,job,po,pv,pvc,cm,secret --all
    ${kubectl} delete sc nsfs-local
}

function noobaa_install {
    #noobaa timeout install # Maybe when creating server we can use local PV
    clean_leftovers
    install
    test_noobaa status
    kuberun get noobaa
    kuberun describe noobaa
    test_admission_deployment
}

function noobaa_install_external {
    #noobaa timeout install # Maybe when creating server we can use local PV
    clean_leftovers
    run_external_postgres
    install_external
    test_noobaa status
    kuberun get noobaa
    kuberun describe noobaa
}

function noobaa_install_external_ssl {
    #noobaa timeout install # Maybe when creating server we can use local PV
    mkdir -p -m 755 certs
	openssl ecparam -name prime256v1 -genkey -noout -out certs/ca.key
	openssl req -new -x509 -sha256 -key certs/ca.key -out certs/ca.crt -subj "/CN=ca.noobaa.com"
    openssl genrsa -out certs/server.key 2048
    openssl req -new -sha256 -key certs/server.key -out certs/server.csr -subj "/CN=postgres-external.${NAMESPACE}.svc"
    openssl x509 -req -in certs/server.csr -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -out certs/server.crt -days 365 -sha256
    openssl ecparam -name prime256v1 -genkey -noout -out certs/client.key
	openssl req -new -sha256 -key certs/client.key -out certs/client.csr -subj "/CN=postgres"
	openssl x509 -req -in certs/client.csr -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -out certs/client.crt -days 365 -sha256
    clean_leftovers
    run_external_postgres_ssl
    install_external_ssl
    test_noobaa status
    kuberun get noobaa
    kuberun describe noobaa
}

function test_admission_deployment {
    kuberun get Secret "admission-webhook-secret"
    kuberun get ValidatingWebhookConfiguration "admission-validation-webhook"
    kuberun get Service "admission-webhook-service"
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

function check_pgdb_config_override {
    local timeout=0
    local temp_file=`echo /tmp/test-$(date +%s).json`
    local current_max_connections=`${kubectl} exec noobaa-db-pg-0 -- psql -c "SELECT MAX(setting) FROM pg_file_settings WHERE name = 'max_connections';" | awk 'NR==3 {print $1}'`
    local final_max_connections=$((current_max_connections + 100))
    printf "{\"spec\":{\"dbConf\":\"\\\nmax_connections = $final_max_connections\"}}" > $temp_file

    kuberun silence patch noobaas.noobaa.io noobaa --patch-file $temp_file --type merge

    while [[ "${final_max_connections}" != "${current_max_connections}" ]]
    do
        echo_time "💬  Waiting for PostgreSQL DB max_connections to match the value specified in dbConf"
        timeout=$((timeout+10))
        sleep 10
        current_max_connections=`${kubectl} exec noobaa-db-pg-0 -- psql -c "SELECT MAX(setting) FROM pg_file_settings WHERE name = 'max_connections';" | awk 'NR==3 {print $1}'`
        if [ ${timeout} -ge 180 ] 
        then
            echo_time "❌  reached the timeout for waiting to the update"
            break
        fi
    done 

    if [[ "${final_max_connections}" == "${current_max_connections}" ]]
    then
        echo_time "✅  PostgreSQL DB config updated successfully"
    else
        echo_time "❌  PostgreSQL DB config didn't got updated, Exiting"
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
    done < <(test_noobaa silence status --show-secrets)
    if [ -z ${AWS_ACCESS_KEY_ID} ] || [ -z ${AWS_SECRET_ACCESS_KEY} ]
    then
        echo_time "❌  Could not get AWS credentials, Exiting"
        exit 1
    fi
    local SECRET=$(dirname ${0})/resources/empty-secret.yaml
    local access_key="  AWS_ACCESS_KEY_ID: ${AWS_ACCESS_KEY_ID}"
    printf "\n${access_key}" >> ${SECRET}
    local secret_key="  AWS_SECRET_ACCESS_KEY: ${AWS_SECRET_ACCESS_KEY}"
    printf "\n${secret_key}" >> ${SECRET}
    kuberun create -f $SECRET
    export SECRET_NAME="empty-secret"
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
            --secret-name ${SECRET_NAME}
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
    local type="s3-compatible"
    local buckets="failns.bucket1"
    local namespacestore="namespacestore.fail"

    test_noobaa bucket create ${buckets}

    # Should fail due to a access/secret key already in use, in case the user didn't want to use it as secret refernce  
    yes n | test_noobaa should_fail namespacestore create ${type} ${namespacestore} \
        --target-bucket ${buckets} \
        --endpoint s3.${NAMESPACE}.svc.cluster.local:443 \
        --access-key ${AWS_ACCESS_KEY_ID} \
        --secret-key ${AWS_SECRET_ACCESS_KEY}

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
    yes | test_noobaa should_fail namespacestore create ${type} ${namespacestore} \
        --fs-backend 'GPFS' \
        --pvc-name ${pvc} \
        --sub-path '/'
    
    #Sub-path contains '..'
    yes | test_noobaa should_fail namespacestore create ${type} ${namespacestore} \
        --fs-backend 'GPFS' \
        --pvc-name ${pvc} \
        --sub-path 'subpath/../'

    #Valid sub-path
    yes | test_noobaa namespacestore create ${type} ${namespacestore} \
        --fs-backend 'GPFS' \
        --pvc-name ${pvc} \
        --sub-path 'subpath'
    
    test_noobaa namespacestore list

    #cleanup
    test_noobaa silence namespacestore delete ${namespacestore}
    kuberun get pv,pvc


    echo_time "✅  namespacestore nsfs validator is done"
}

function check_pv_pool_resources {
    echo_time "💬  Staring PV Pool resources cycle"

    # Minimum CPU     100m
    #         Memory  400Mi
    test_noobaa should_fail backingstore create pv-pool request-small-cpu \
            --num-volumes 1 \
            --pv-size-gb 16 \
            --request-cpu 50m

    test_noobaa should_fail backingstore create pv-pool request-small-memory \
            --num-volumes 1 \
            --pv-size-gb 16 \
            --request-memory 100Mi

    test_noobaa should_fail backingstore create pv-pool request-larger-limit \
            --num-volumes 1 \
            --pv-size-gb 16 \
            --request-cpu 300m \
            --limit-cpu 200m

    local mem=400
    local cpu=100
    if [ "$RESOURCE" == "dev" ]
    then
        mem=500
        cpu=500
    fi
    test_noobaa backingstore create pv-pool minimum-request-limit \
            --num-volumes 1 \
            --pv-size-gb 16 \
            --request-cpu $(cpu)m \
            --request-memory $(mem)Mi \
            --limit-cpu $(cpu)m \
            --limit-memory $(mem)Mi
    #TOD see why it fails, currently disabling as it takes 10 mins.
    # time="2022-04-11T14:18:17Z" level=error msg="❌ BackingStore \"large-request-limit\" Phase is \"Rejected\": Failed connecting all pods in backingstore for more than 10 minutes Current failing: 1 from requested: 1"
    # NAME                           TYPE      TARGET-BUCKET   PHASE      AGE      
    # large-request-limit            pv-pool                   Rejected   10m7s    
    # test_noobaa backingstore create pv-pool large-request-limit \
    #         --num-volumes 1 \
    #         --pv-size-gb 16 \
    #         --request-cpu 300m \
    #         --request-memory 500Mi \
    #         --limit-cpu 400m \
    #         --limit-memory 600Mi

    test_noobaa backingstore list
    test_noobaa status
    kuberun get backingstore
    kuberun describe backingstore

    test_noobaa backingstore delete minimum-request-limit
    # test_noobaa backingstore delete large-request-limit

    echo_time "✅  PV Pool resources cycle is done"
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
            --secret-name ${SECRET_NAME}
        wait_for_backingstore_ready ${backingstore[cycle]}
    done
    test_noobaa backingstore list
    test_noobaa status
    kuberun get backingstore
    kuberun describe backingstore
    check_S3_compatible_validator
    echo_time "✅  s3 compatible cycle is done"
}

function check_S3_compatible_validator {
    local type="s3-compatible"
    local buckets="fails3.bucket"
    local backingstore="fail.compatible1"

    test_noobaa bucket create ${buckets}

    # Should fail due to a access/secret key already in use, in case the user didn't want to use it as secret refernce 
    yes n | test_noobaa should_fail backingstore create ${type} ${backingstore} \
        --target-bucket ${buckets} \
        --endpoint s3.${NAMESPACE}.svc.cluster.local:443 \
        --access-key ${AWS_ACCESS_KEY_ID} \
        --secret-key ${AWS_SECRET_ACCESS_KEY}
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
            --secret-name ${SECRET_NAME}
        test_noobaa backingstore status ${backingstore[cycle]}
    done
    test_noobaa backingstore list
    test_noobaa status
    kuberun get backingstore
    kuberun describe backingstore
    check_IBM_cos_validator
    echo_time "✅  ibm cos cycle is done"
}

function check_IBM_cos_validator {
    local type="ibm-cos"
    local buckets="failIBM.bucket"
    local backingstore="fail.ibmcos"

    test_noobaa bucket create ${buckets}

    # Should fail due to a access/secret key already in use, in case the user didn't want to use it as secret refernce 
    yes n | test_noobaa should_fail backingstore create ${type} ${backingstore} \
        --target-bucket ${buckets} \
        --endpoint s3.${NAMESPACE}.svc.cluster.local:443 \
        --access-key ${AWS_ACCESS_KEY_ID} \
        --secret-key ${AWS_SECRET_ACCESS_KEY}
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

function bz_2038884 {
    test_noobaa bucketclass create placement-bucketclass testbucketclass --backingstores=noobaa-default-backing-store
    test_noobaa obc create --bucketclass=testbucketclass testobc
    test_noobaa bucketclass delete testbucketclass
    test_noobaa obc delete testobc
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

function account_cycle {
    echo_time "💬  Starting the account cycle"
    local buckets=($(test_noobaa silence bucket list  | grep -v "BUCKET-NAME" | awk '{print $1}'))
    local backingstores=($(test_noobaa silence backingstore list | grep -v "NAME" | awk '{print $1}'))
    test_noobaa account create account1  #default_resource should be the system default
    test_noobaa account create account2 --default_resource ${backingstores[0]}

    #admin account that have a secret but no CRD 
    account_regenerate_keys "admin@noobaa.io"
    account_update_keys "admin@noobaa.io" "Aa123456789123456789" "Aa+/123456789123456789123456789123456789"
    #admin account that don't have a secret and don't have CRD 
    account_regenerate_keys "operator@noobaa.io"
    account_update_keys "operator@noobaa.io" "Ab987654321987654321" "Ab+/987654321987654321987654321987654321"
    # testing account reset password
    account_reset_password "admin@noobaa.io"
    # testing nsfs accounts
    # account_nsfs_cycle TODO re-enable.
    # update account default resource
    account_update
    # update crd account default resource
    account_update "crd"
    echo_time "✅  noobaa account cycle is done"
}

function account_update {
    local crd=${1}
    local new_default_resource="backing-store-upd"
    if [ "$(kuberun get backingstore | grep -w ${new_default_resource} | wc -l)" != "1" ]
    then
        echo_time "💬  Creating backingstore ${new_default_resource}"
        test_noobaa backingstore create pv-pool ${new_default_resource} --num-volumes 1 --pv-size-gb 16
    fi
    if [ -z ${crd} ]
    then
        echo_time "💬  Checking account update cycle"
        local account_name=$(test_noobaa api account list_accounts {} -o json | jq -r '.accounts[0].email')
        local default_resource=$(test_noobaa api account list_accounts {} -o json | jq -r '.accounts[0].default_resource')
        echo_time "💬  Updating ${new_default_resource} as default_resource for noobaa account"
        test_noobaa account update ${account_name} --new_default_resource=${new_default_resource}
        local curr_default_resource=$(test_noobaa api account list_accounts {} -o json | jq -r '.accounts[0].default_resource')
        if [ "${new_default_resource}" != "${curr_default_resource}" ]
        then
            echo_time "❌ Looks like account not updated, Exiting"
            exit 1
        else
            echo_time "✅  Update account default_resource successful"
            test_noobaa account update ${account_name} --new_default_resource=${default_resource}
            curr_default_resource=$(test_noobaa api account list_accounts {} -o json | jq -r '.accounts[0].default_resource')
            if [ "${default_resource}" != "${curr_default_resource}" ]
            then
                echo_time "❌ Looks like account not updated to default state, Exiting"
                exit 1
            fi
        fi
    else
        echo_time "💬  Checking crd based account update cycle"
        local account_name="test-account"
        echo_time "💬  Creating crd account ${account_name}"
        test_noobaa account create ${account_name}
        local default_resource=$(kuberun get NoobaaAccount ${account_name} -n test -o json | jq -r '.spec.default_resource')
        echo_time "💬  Updating ${new_default_resource} as default_resource for noobaa account"
        test_noobaa account update ${account_name} --new_default_resource=${new_default_resource}
        local curr_default_resource=$(kuberun get NoobaaAccount ${account_name} -n test -o json | jq -r '.spec.default_resource')
        if [ "${new_default_resource}" != "${curr_default_resource}" ]
        then
            echo_time "❌ Looks like crd account not updated, Exiting"
            exit 1
        else
            echo_time "✅  Update crd account default_resource successful"
            echo_time "💬  Deleting crd account ${account_name}"
            test_noobaa account delete ${account_name}
            echo_time "💬  Deleting backingstore ${new_default_resource}"
            test_noobaa backingstore delete ${new_default_resource}
        fi
    fi
}

function account_regenerate_keys {
    local account=${1}
    local AWS_ACCESS_KEY_ID
    local AWS_SECRET_ACCESS_KEY
    if [ "${account}" != "operator@noobaa.io" ]
    then
        while read line
        do
            if [[ ${line} =~ (AWS_ACCESS_KEY_ID|AWS_SECRET_ACCESS_KEY) ]]
            then
                eval $(echo ${line//\"/} | sed -e 's/ //g' -e 's/:/=/g')
            fi
        done < <(test_noobaa account status ${account} --show-secrets)
    fi

    local ACCESS_KEY_ID_before=${AWS_ACCESS_KEY_ID}
    local SECRET_ACCESS_KEY_before=${AWS_SECRET_ACCESS_KEY}
    while read line
    do
        if [[ ${line} =~ (AWS_ACCESS_KEY_ID|AWS_SECRET_ACCESS_KEY) ]]
        then
            eval $(echo ${line//\"/} | sed -e 's/ //g' -e 's/:/=/g')
        fi
    done < <(yes | test_noobaa account regenerate ${account} --show-secrets)

    if [ "${AWS_ACCESS_KEY_ID}" == "${ACCESS_KEY_ID_before}" ]
    then
        echo_time "❌ Looks like the ACCESS_KEY were not regenerated, Exiting"
        exit 1
    fi

    if [ "${AWS_SECRET_ACCESS_KEY}" == "${SECRET_ACCESS_KEY_before}" ]
    then
        echo_time "❌ Looks like the SECRET_ACCESS were not regenerated, Exiting"
        exit 1
    fi
}

function account_update_keys {
    local account=${1}
    local access_key=${2}
    local secret_key=${3}

    local AWS_ACCESS_KEY_ID
    local AWS_SECRET_ACCESS_KEY
    if [ "${account}" != "operator@noobaa.io" ]
    then
        while read line
        do
            if [[ ${line} =~ (AWS_ACCESS_KEY_ID|AWS_SECRET_ACCESS_KEY) ]]
            then
                eval $(echo ${line//\"/} | sed -e 's/ //g' -e 's/:/=/g')
            fi
        done < <(test_noobaa account status ${account} --show-secrets)
    fi

    # should fail if account access key not matching the criteria
    test_noobaa should_fail account credentials ${account} --access-key="Afjlkdsfnla" --secret-key="Aa+/123456789123456789123456789123456789"
    # should fail if account secret key is not matching the criteria
    test_noobaa should_fail account credentials ${account} --access-key="Aa123456789123456789" --secret-key="Aandlfknslkdnf"

    local ACCESS_KEY_ID_before=${AWS_ACCESS_KEY_ID}
    local SECRET_ACCESS_KEY_before=${AWS_SECRET_ACCESS_KEY}
    while read line
    do
        if [[ ${line} =~ (AWS_ACCESS_KEY_ID|AWS_SECRET_ACCESS_KEY) ]]
        then
            eval $(echo ${line//\"/} | sed -e 's/ //g' -e 's/:/=/g')
        fi
    done < <(yes | test_noobaa account credentials ${account} \
        --access-key=${access_key} --secret-key=${secret_key} --show-secrets)

    if [ "${AWS_ACCESS_KEY_ID}" == "${ACCESS_KEY_ID_before}" ]
    then
        echo_time "❌ Looks like the ACCESS_KEY were not updated, Exiting"
        exit 1
    fi

    if [ "${AWS_SECRET_ACCESS_KEY}" == "${SECRET_ACCESS_KEY_before}" ]
    then
        echo_time "❌ Looks like the SECRET_ACCESS were not updated, Exiting"
        exit 1
    fi
}

function account_reset_password {
    local account=${1}
    local password=$(get_admin_password)
    #reset password should work
    test_noobaa account passwd ${account} --old-password ${password} --new-password "test" --retype-new-password "test"
    # Should fail if the old password is not correct
    test_noobaa should_fail account passwd ${account} --old-password "test1" --new-password "test" --retype-new-password "test"
    # Should fail if we got the same password as the old one
    test_noobaa should_fail account passwd ${account} --old-password "test" --new-password "test" --retype-new-password "test"
    # Should fail if we got the same password twice 
    test_noobaa should_fail account passwd ${account} --old-password "test" --new-password "test1" --retype-new-password "test2"
}

function get_admin_password {
    local password
    while read line
    do
        if [[ ${line} =~ "password" ]]
        then
            password=$(echo ${line//\"/} | awk -F ":" '{print $2}')
        fi
    done < <(yes | test_noobaa status --show-secrets)
    echo ${password}
}

function account_nsfs_cycle {
    local default_resource="fs1"
    # Creating namespacestore to use by the account 
    yes | test_noobaa namespacestore create nsfs ${default_resource} --pvc-name nsfs-vol --fs-backend GPFS
    # Testing that we can create account using namespacestore
    test_noobaa account create fsaccount1 --full_permission --default_resource ${default_resource} --nsfs_account_config --uid 123 --gid 456
    # should fail if the default_resource does not exists
    test_noobaa should_fail account create fsaccount2 --full_permission --default_resource not_exists --nsfs_account_config --uid 123 --gid 456
    # should fail if the uid is not a number   
    test_noobaa should_fail account create fsaccount3 --full_permission --default_resource ${default_resource} --nsfs_account_config --uid fail --gid 456
    # should fail if the gid is not a number
    test_noobaa should_fail account create fsaccount4 --full_permission --default_resource ${default_resource} --nsfs_account_config --uid 123 --gid fail
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

function delete_account {
    local accounts=($(test_noobaa silence account list | grep -v "NAME" | awk '{print $1}'))
    for account in ${accounts[@]}
    do
        test_noobaa account delete ${account}
    done
    echo_time "✅  delete accounts is done"
}

function delete_non_existing_resources {
    test_noobaa should_fail obc delete non-existing-obc
    test_noobaa should_fail bucketclass delete non-existing-bc
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
    delete_accounts
    delete_non_existing_resources
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
    echo "{\"rules\":[{ \"rule_id\": \"rule-1\", \"destination_bucket\": \"first.bucket\", \"filter\": {\"prefix\": \"d\"}} ]}" > replication1.json
    echo "{\"rules\":[{ \"rule_id\": \"rule-2\", \"destination_bucket\": \"first.bucket\", \"filter\": {\"prefix\": \"e\"}} ]}" > replication2.json
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

function check_default_backingstore {
    echo_time "💬 Checking if Noobaa Default Backingstore is already present"
    local default_backing=$(kuberun get backingstore | grep -w noobaa-default-backing-store | wc -l)
    if [[ "${default_backing}" =~ "1" ]]
    then
        echo_time "✅  Default Backingstore is already present"
    else
        echo_time "❌  Default Backingstore is not already present, Exiting"
        exit 1
    fi

    echo_time "💬 Disabling Noobaa default backingstore"
	kuberun patch noobaa/noobaa --type json --patch='[{"op":"add","path":"/spec/manualDefaultBackingStore","value":true}]'

    echo_time "💬 Deleting Noobaa default backingstore and its connected instances"
    echo_time "💬 Deleting buckets"
    NOOBAA_ACCESS_KEY=$(kuberun get secret noobaa-admin -n test -o json | jq -r '.data.AWS_ACCESS_KEY_ID|@base64d')
    NOOBAA_SECRET_KEY=$(kuberun get secret noobaa-admin -n test -o json | jq -r '.data.AWS_SECRET_ACCESS_KEY|@base64d')
    ENDPOINT=$(kuberun get noobaa noobaa -n test -o json | jq -r '.status.services.serviceS3.nodePorts[0]')
    echo $ENDPOINT
    AWS_ACCESS_KEY_ID=$NOOBAA_ACCESS_KEY AWS_SECRET_ACCESS_KEY=$NOOBAA_SECRET_KEY AWS_EC2_METADATA_DISABLED=true aws --endpoint $ENDPOINT --no-verify-ssl s3 ls
    for bucket in $(AWS_ACCESS_KEY_ID=$NOOBAA_ACCESS_KEY AWS_SECRET_ACCESS_KEY=$NOOBAA_SECRET_KEY AWS_EC2_METADATA_DISABLED=true aws --endpoint $ENDPOINT --no-verify-ssl s3 ls | awk '{print $3}'); 
    do  
        AWS_ACCESS_KEY_ID=$NOOBAA_ACCESS_KEY AWS_SECRET_ACCESS_KEY=$NOOBAA_SECRET_KEY AWS_EC2_METADATA_DISABLED=true aws --endpoint $ENDPOINT --no-verify-ssl s3 rb "s3://${bucket}" --force ; 
    done
    "💬 Deleting non-default accounts"
    delete_account

    echo_time "💬 Creating new-default-backing-store and updating the admin account default_resourse with it"
    test_noobaa backingstore create pv-pool new-default-backing-store --num-volumes 1 --pv-size-gb 16
    test_noobaa account update admin@noobaa.io --new_default_resource=new-default-backing-store
    test_noobaa api account list_accounts {}
    test_noobaa account list

    echo_time "💬 Deleting backingstore noobaa-default-backing-store"
    kuberun delete backingstore noobaa-default-backing-store -n test | kubectl patch -n test backingstore/noobaa-default-backing-store --type json --patch='[ { "op": "remove", "path": "/metadata/finalizers" } ]'

    default_backing=$(kuberun get backingstore | grep -w noobaa-default-backing-store | wc -l)
    while [[ "${default_backing}" =~ "1" ]]
    do
        echo_time "💬  Waiting for default backingstore to be deleted"
        sleep 3
        default_backing=$(kuberun get backingstore | grep -w noobaa-default-backing-store | wc -l)
    done
    sleep 20

    echo_time "💬 Checking if Noobaa Default Backingstore is Reconciled"
    local default_backing=$(kuberun get backingstore | grep -w noobaa-default-backing-store | wc -l)
    if [[ "${default_backing}" =~ "0" ]]
    then
        echo_time "✅  Default Backingstore is not reconciled, Successful"
    else
        echo_time "❌  Default Backingstore is reconciled, Exiting"
        exit 1
    fi

    echo_time "💬 Enabling Noobaa default backingstore"
    kuberun patch noobaa/noobaa --type json --patch='[{"op":"add","path":"/spec/manualDefaultBackingStore","value":false}]'
    sleep 10s
}

function check_dbdump {
    echo_time "💬  Generating db dump"

    # Generate db dump at /tmp/<random_dir>
    rand_dir=`tr -dc A-Za-z0-9 </dev/urandom | head -c 13 ; echo ''`
    mkdir /tmp/$rand_dir
    test_noobaa db-dump --dir /tmp/$rand_dir

    # Check whether dump was created
    dump_file_name=`ls -l /tmp/$rand_dir | grep noobaa_db_dump | awk '{ print $9 }'`
    if [ ! -f "/tmp/$rand_dir/$dump_file_name" ]
    then
        echo_time "❌  db dump was not generated"
        exit 1
    fi

    # Remove dump file
    rm /tmp/$rand_dir/$dump_file_name

    # Generate db dump through diagnostics API
    echo_time "💬  Generating db dump through diagnostics"
    test_noobaa diagnostics collect --db-dump --dir /tmp/$rand_dir

    # Check whether dump was created
    diagnose_file_name=`ls -l /tmp/$rand_dir | grep noobaa_diagnostics | awk '{ print $9 }'`
    dump_file_name=`ls -l /tmp/$rand_dir | grep noobaa_db_dump | awk '{ print $9 }'`
    if [ ! -f "/tmp/$rand_dir/$dump_file_name" ]
    then
        echo_time "❌  db dump was not generated"
        exit 1
    fi

    # Remove diagnostics and dump files
    rm -rf /tmp/$rand_dir
}

function test_noobaa_cr_deletion() {
    local resp
    resp=$(kubectl -n ${NAMESPACE} delete noobaas.noobaa.io noobaa 2>&1 >/dev/null)
    if [ $? -ne 0 ]; then
        echo $resp
        if [[ $resp == *"Noobaa cleanup policy is not set, blocking Noobaa deletion"* ]]; then
            echo_time "✅  Noobaa CR deletion test passed"
        else
            echo_time "❌  Noobaa CR deletion test failed"
            exit 1
        fi
    else
        echo_time "❌  Noobaa CR deletion test failed: kubectl delete returned 0"
        exit 1
    fi
}

function test_noobaa_loadbalancer_source_subnet() {
    local timeout=0
    local temp_file=`echo /tmp/test-$(date +%s).json`
    local subnet1=10.0.0.0/16
    local subnet2=172.18.0.0/32
    cat <<EOF > $temp_file
{
    "spec": {
        "loadBalancerSourceSubnets":  {
            "s3": ["$subnet1"],
            "sts": ["$subnet2"]
        }   
    }
}
EOF

    kuberun silence patch noobaas.noobaa.io noobaa --patch-file $temp_file --type merge

    while [ $timeout -lt 60 ]; do
        sleep 1
        timeout=$((timeout+1))
        if [ $timeout -eq 60 ]; then
            echo_time "❌  Noobaa loadbalancer source subnet test failed"
            exit 1
        fi

        local passed=true

        local loadBalancerSourceRanges=`kubectl get services s3 -n ${NAMESPACE} -o json | jq -rc '.spec.loadBalancerSourceRanges'`
        if [ "$loadBalancerSourceRanges" == "[\"$subnet1\"]" ]; then
            echo_time "✅  Noobaa loadbalancer source subnet verified for service s3"
        else
            echo_time "❌  Noobaa loadbalancer source subnet test failed for service s3"
            passed=false
        fi

        local loadBalancerSourceRanges=`kubectl get services sts -n ${NAMESPACE} -o json | jq -rc '.spec.loadBalancerSourceRanges'`
        if [ "$loadBalancerSourceRanges" == "[\"$subnet2\"]" ]; then
            echo_time "✅  Noobaa loadbalancer source subnet verified for service sts"
        else
            echo_time "❌  Noobaa loadbalancer source subnet test failed for service sts"
            passed=false
        fi

        if [ "$passed" == "true" ]; then
            echo_time "✅  Noobaa loadbalancer source subnet test passed"
            break
        fi
    done
}

function test_multinamespace_bucketclass() {

    # Helper function to create and test bucketclass
    function test_create_bucketclass() {
        local timeout=0
        local fail_time=600
        local bucketclass_name=$1
        local backingstore_name=$2
        local namespace=$3
        local provisioner=$4
        local fail=$5

        cat <<EOF | kubectl -n $namespace apply -f -
apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
    name: $bucketclass_name
    labels:
        noobaa-operator: $provisioner
spec:
    placementPolicy:
        tiers:
        - backingStores:
          - $backingstore_name
EOF

        # If fail is set to true then expect the test to fail
        if [ "$fail" == "true" ]; then
            local bucketclass=`kubectl -n $namespace get bucketclass $bucketclass_name -o=go-template='{{.status}}'`
            if [ "$bucketclass" == "<no value>" ]; then
                echo_time "✅  [${FUNCNAME[0]}]: Noobaa bucketclass creation - not picked by the operator - test passed"
            else
                echo_time "❌  [${FUNCNAME[0]}]: Noobaa bucketclass creation - picked by the operator - test failed"
                exit 1
            fi
        else
            while [ $timeout -lt $fail_time ]; do
                sleep 1
                timeout=$((timeout+1))
                if [ $timeout -eq $fail_time ]; then
                    echo_time "❌  [${FUNCNAME[0]}]: Noobaa multinamespace bucketclass test failed"
                    exit 1
                fi

                local bucketclass=`kubectl -n $namespace get bucketclass $bucketclass_name -o=go-template='{{.status.phase}}'`
                if [ "$bucketclass" == "" ]; then
                    echo_time "❌  [${FUNCNAME[0]}]: Noobaa multinamespace bucketclass test failed for bucketclass $bucketclass_name"
                elif [ "$bucketclass" == "Ready" ]; then
                    echo_time "✅  [${FUNCNAME[0]}]: Noobaa multinamespace bucketclass verified for bucketclass $bucketclass_name"
                    break
                else
                    echo_time "❌  [${FUNCNAME[0]}]: Noobaa multinamespace bucketclass test failed for bucketclass $bucketclass_name - status is $bucketclass"
                fi
            done
        fi
    }


    # Helper function to create OBC
    function test_create_obc() {
        local timeout=0
        local obc_name=$1
        local bucketclass_name=$2
        local namespace=$3

        cat <<EOF | kubectl -n $namespace apply -f -
apiVersion: objectbucket.io/v1alpha1
kind: ObjectBucketClaim
metadata:
    name: $obc_name
spec:
    bucketName: $obc_name
    storageClassName: ${NAMESPACE}.noobaa.io
    additionalConfig:
        bucketclass: $bucketclass_name
EOF

        while [ $timeout -lt 600 ]; do
            sleep 1
            timeout=$((timeout+1))
            if [ $timeout -eq 600 ]; then
                echo_time "❌  [${FUNCNAME[0]}]: Noobaa multinamespace bucketclass test failed"
                exit 1
            fi


            local obc=`kubectl -n $namespace get obc $obc_name -o=go-template='{{.status.phase}}'`
            if [ "$obc" == "" ]; then
                echo_time "❌  [${FUNCNAME[0]}]: Noobaa multinamespace bucketclass test failed for obc $obc_name"
            elif [ "$obc" == "Bound" ]; then
                echo_time "✅  [${FUNCNAME[0]}]: Noobaa multinamespace bucketclass verified for obc $obc_name"
                break
            else
                echo_time "❌  [${FUNCNAME[0]}]: Noobaa multinamespace bucketclass test failed for obc $obc_name - status is $obc"
            fi
        done
    }

    # Helper function to delete OBC
    function test_delete_obc() {
        local timeout=0
        local obc_name=$1
        local namespace=$2

        kubectl -n $namespace delete obc $obc_name
    }

    # Helper function to delete bucketclass
    function test_delete_bucketclass() {
        local timeout=0
        local bucketclass_name=$1
        local namespace=$2

        kubectl -n $namespace delete bucketclass $bucketclass_name
    }

    # Test multinamespace bucketclass - system namespace
    function test_multinamespace_bucketclass_system_namespace() {
        test_create_bucketclass multinamespace-bucketclass noobaa-default-backing-store ${NAMESPACE} ${NAMESPACE}
        test_create_obc multinamespace-obc multinamespace-bucketclass ${NAMESPACE}

        test_delete_obc multinamespace-obc ${NAMESPACE}
        test_delete_bucketclass multinamespace-bucketclass ${NAMESPACE}

        echo_time "✅  [${FUNCNAME[0]}]: Noobaa multinamespace bucketclass test passed"
    }

    # Test multinamespace bucketclass - non-system namespace correct operator
    function test_multinamespace_bucketclass_non_system_namespace_correct_operator() {
        local random_namespace=`echo test-ns-$(date +%s)`

        kubectl create namespace $random_namespace

        test_create_bucketclass multinamespace-bucketclass noobaa-default-backing-store ${random_namespace} ${NAMESPACE}
        test_create_obc multinamespace-obc multinamespace-bucketclass ${random_namespace}

        test_delete_obc multinamespace-obc ${random_namespace}
        test_delete_bucketclass multinamespace-bucketclass ${random_namespace}

        kubectl delete namespace $random_namespace

        echo_time "✅  [${FUNCNAME[0]}]: Noobaa multinamespace bucketclass test passed"
    }

    # Test multinamespace bucketclass - non-system namespace incorrect operator
    function test_multinamespace_bucketclass_non_system_namespace_incorrect_operator() {
        local random_namespace=`echo test-ns-$(date +%s)`
        local random_operator=`echo test-op-$(date +%s)`

        kubectl create namespace $random_namespace

        test_create_bucketclass multinamespace-bucketclass noobaa-default-backing-store ${random_namespace} ${random_operator} "true"
        test_delete_bucketclass multinamespace-bucketclass ${random_namespace}

        kubectl delete namespace $random_namespace

        echo_time "✅  [${FUNCNAME[0]}]: Noobaa multinamespace bucketclass test passed"
    }

    test_multinamespace_bucketclass_system_namespace
    test_multinamespace_bucketclass_non_system_namespace_correct_operator
    test_multinamespace_bucketclass_non_system_namespace_incorrect_operator

    echo_time "✅  [${FUNCNAME[0]}]: Noobaa multinamespace bucketclass test passed"
}

# Should fail due to:
function obc_nsfs_negative_tests {
    # No GID
    yes n | test_noobaa should_fail obc create testobc --uid 42
    # No UID
    yes n | test_noobaa should_fail obc create testobc --gid 505
    # Distingusihed name provided in conjunction with UID
    yes n | test_noobaa should_fail obc create testobc --distinguished-name 'test' --uid 42
    # GID and UID (valid) but no bucketclass
    yes n | test_noobaa should_fail obc create testobc --uid 42 --gid 505
}

function test_create_obc_with_nsfs_acc_cfg_uid_gid_logic {
    local uid=$1
    local gid=$2
    local obc_name="giduidtestobc"
    test_noobaa obc create ${obc_name} --uid ${uid} --gid ${gid} --bucketclass noobaa-default-bucket-class
    local obc_account_nsfs_account=$(test_noobaa api account list_accounts {} -ojson | jq '.accounts[] | select(.bucket_claim_owner | test("'${obc_name}'"))?' | jq '.nsfs_account_config')
    local account_gid=$(echo $obc_account_nsfs_account | jq '.gid')
    local account_uid=$(echo $obc_account_nsfs_account | jq '.uid')
    test_noobaa obc delete ${obc_name}
    if [[ $account_gid != $gid || $account_uid != $uid ]]; then
        echo_time "❌  [${FUNCNAME[0]}]: Noobaa obc nsfs account creation test failed. Expected: uid=$uid, gid=$gid, Got: uid=$account_uid, gid=$account_gid"
        exit 1
    fi
}

# test_create_obc_with_nsfs_acc_cfg_uid_gid but it iterates over several pairs of UID and GID - [0,0], [42, 505]
function test_create_obc_with_nsfs_acc_cfg_uid_gid {
    for uid in 0 42; do
        for gid in 0 505; do
            test_create_obc_with_nsfs_acc_cfg_uid_gid_logic $uid $gid
        done
    done
}

function test_create_obc_with_nsfs_acc_distinguished_name {
    local distinguished_name="someuser"
    local obc_name="distinguishednameobc"
    test_noobaa obc create ${obc_name} --distinguished-name $distinguished_name --bucketclass noobaa-default-bucket-class
    local obc_account_nsfs_account=$(test_noobaa api account list_accounts {} -ojson | jq '.accounts[] | select(.bucket_claim_owner | test("'${obc_name}'"))?' | jq '.nsfs_account_config')
    local account_distinguished_name=$(echo $obc_account_nsfs_account | jq -r '.distinguished_name')
    test_noobaa obc delete ${obc_name}
    if [[ $account_distinguished_name != $distinguished_name ]]; then
        echo_time "❌  [${FUNCNAME[0]}]: Noobaa obc nsfs account creation test failed. Expected: distinguished_name=$distinguished_name, Got: distinguished_name=$account_distinguished_name"
        exit 1
    fi
}