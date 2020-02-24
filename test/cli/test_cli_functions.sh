#!/bin/bash

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
    local rc
    local run_with_timeout=false
    local should_fail=false
    local silence=false
    if [ "${1}" == "failure" ]
    then
        should_fail=true
        shift
    fi
    if [ "${1}" == "timeout" ]
    then
        run_with_timeout=true
        shift
    fi
    if [ "${1}" == "silence" ]
    then
        silence=true
        shift
    fi
    local options=$*

    if [ -z "${noobaa}" ]
    then
        echo_time "❌  The noobaa variable must be define in the shell"
        exit 1
    fi
    if ${run_with_timeout}
    then
        ${noobaa} ${options} &
        PID=$!
        # We are trapping SIGHUP and SIGINT for clean exit.
        trap "clean ${PID}" 1 2
        # When we are running with timeout because the command runs in the background
        timeout ${PID} ${options}
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
    local PID=${1}
    shift
    local options=$*
    local START_TIME=${SECONDS}

    if [ -z "${TIMEOUT}" ]
    then
        cho "❌  The TIMEOUT variable must be define in the shell"
        exit 1
    fi

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
            echo_time "❌  ${noobaa} ${options} reached timeout, Exiting"
            exit 1
        fi
    done
}

function install {
    test_noobaa install --mini
    local status=$(kuberun get noobaa noobaa -o json | jq -r '.status.phase' 2> /dev/null)
    while [ "${status}" != "Ready" ]
    do
        echo_time "Waiting for status Ready, Status is ${status}"
        sleep 10
        status=$(kuberun get noobaa noobaa -o json | jq -r '.status.phase' 2> /dev/null)
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
    echo_time "Staring compatible cycle"
    local cycle
    local type="s3-compatible"
    local buckets=("first.bucket" "second.bucket")
    local backingstore=("compatible1" "compatible2")

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
    echo_time "✅  s3 compatible cycle is done"
}

function check_IBM_cos {
    echo_time "Staring IBM cos cycle"
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
    echo_time "Starting the bucketclass cycle"
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

function obc_cycle {
    echo_time "Starting the obc cycle"
    local bucket
    local buckets=()

    local bucketclass_list_array=($(test_noobaa silence bucketclass list | awk '{print $1}' | grep -v NAME | grep -v noobaa-default-bucket-class))
    for bucketclass in ${bucketclass_list_array[@]}
    do
        buckets+=("bucket${bucketclass//[a-zA-Z.-]/}")
        if [ "${bucketclass//[a-zA-Z.-]/}" == "3" ]
        then
            flag="--app-namespace default"
        fi
        test_noobaa timeout obc create ${buckets[$((${#buckets[@]}-1))]} --bucketclass ${bucketclass} ${flag}
        unset flag
    done
    test_noobaa obc list
    for bucket in ${buckets[@]}
    do
        test_noobaa timeout obc status ${bucket}
    done
    kuberun get obc
    kuberun describe obc
    kuberun get obc,ob,secret,cm -l noobaa-obc

    # aws s3 --endpoint-url XXX ls
    echo_time "✅  obc cycle is done"
}

function delete_backingstore_path {
    local object_bucket backing_store
    local backingstore=($(test_noobaa silence backingstore list | grep -v "NAME" | awk '{print $1}'))
    local bucketclass=($(test_noobaa silence bucketclass list  | grep ${backingstore[1]} | awk '{print $1}'))
    local obc=($(test_noobaa silence obc list | grep -v "BUCKET-NAME" | awk '{print $2}'))
    echo_time "Starting the delete related ${backingstore[1]} paths"

    test_noobaa failure backingstore delete ${backingstore[1]}
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
    test_noobaa failure backingstore delete ${backingstore[0]}
    echo_time "✅  delete ${backingstore[1]} path is done"
}

function check_deletes {
    echo_time "Starting the delete cycle"
    local obc=($(test_noobaa silence obc list | grep -v "NAME\|default" | awk '{print $2}'))
    local bucketclass=($(test_noobaa silence bucketclass list  | grep -v NAME | awk '{print $1}'))
    local backingstore=($(test_noobaa silence backingstore list | grep -v "NAME" | awk '{print $1}'))
    test_noobaa obc delete ${obc[0]}
    test_noobaa bucketclass delete ${bucketclass[0]}
    test_noobaa backingstore list
    delete_backingstore_path
    echo_time "✅  delete cycle is done"
}
