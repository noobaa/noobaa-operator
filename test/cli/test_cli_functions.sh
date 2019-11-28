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
        echo "❌  The kubectl variable must be define in the shell"
        exit 1
    fi
    ${kubectl} ${options}
    if [ $? -ne 0 ]
    then 
        echo "❌  ${kubectl} ${options} failed, Exiting"
        exit 1
    elif [ ! ${silence} ]
    then
        echo "✅  ${kubectl} ${options} passed" 
    fi
}

function noobaa {
    local rc
    local run_with_timeout=false
    local silence=false
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
        echo "❌  The noobaa variable must be define in the shell"
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
            echo "❌  ${noobaa} ${options} failed, Exiting"
            exit 1
        elif [ ! ${silence} ]
        then
            echo "✅  ${noobaa} ${options} passed" 
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
            echo "✅  ${noobaa} ${options} passed" 
            break
        fi

        if [ $((START_TIME+TIMEOUT)) -gt ${SECONDS} ]
        then
            sleep 5
        else
            kill -9 ${PID}            
            echo "❌  ${noobaa} ${options} reached timeout, Exiting"
            exit 1
        fi
    done
}

function install {
# TODO: once we can control the core resource amount we can use it,
#       for now we will mimic the install command 
    noobaa crd create
    noobaa system create
    kuberun patch noobaa noobaa --type merge -p '{"spec":{"coreResources":{"requests":{"cpu":"1m","memory":"256Mi"}}}}'
    noobaa operator install
    local status=$(kuberun get noobaa noobaa -o json | jq -r '.status.phase' 2> /dev/null)
    while [ "${status}" != "Ready" ] 
    do 
        echo "Waiting for status Ready, Status is ${status}"
        sleep 10
        status=$(kuberun get noobaa noobaa -o json | jq -r '.status.phase' 2> /dev/null)
    done
}

function noobaa_install {
    #noobaa timeout install # Maybe when creating server we can use local PV
    install
    noobaa status
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
    done < <(noobaa silence status)
    if [ -z ${AWS_ACCESS_KEY_ID} ] || [ -z ${AWS_SECRET_ACCESS_KEY} ]
    then
        echo "❌  Could not get AWS credentials, Exiting"
        exit 1
    fi 
}

function check_S3_compatible {
    echo "Staring compatible cycle"
    local cycle
    local type="s3-compatible"
    local buckets=("first.bucket" "second.bucket")
    local backingstore=("compatible1" "compatible2")

    noobaa bucket create ${buckets[1]}
    for (( cycle=0 ; cycle < ${#backingstore[@]} ; cycle++ ))
    do
        noobaa backingstore create ${type} ${backingstore[cycle]} \
            --target-bucket ${buckets[cycle]} \
            --endpoint 127.0.0.1:6443 \
            --access-key ${AWS_ACCESS_KEY_ID} \
            --secret-key ${AWS_SECRET_ACCESS_KEY}
        noobaa backingstore status ${backingstore[cycle]}
    done
    noobaa backingstore list
    noobaa status
    kuberun get backingstore
    kuberun describe backingstore
    echo "✅  s3 compatible cycle is done"
}

function check_aws_S3 {
    return
    # noobaa bucket create second.bucket
    # noobaa backingstore create aws1 --type aws-s3 --bucket-name znoobaa --access-key XXX --secret-key YYY
    # noobaa backingstore create aws2 --type aws-s3 --bucket-name noobaa-qa --access-key XXX --secret-key YYY
    # noobaa backingstore status aws1
    # noobaa backingstore status aws2
    # noobaa backingstore list
    # noobaa status
    # kubectl get backingstore
    # kubectl describe backingstore
}

function bucketclass_cycle {
    echo "Starting the bucketclass cycle"
    local bucketclass
    local bucketclass_names=()
    local backingstore=()
    local number_of_backingstores=4

    for (( number=0 ; number < number_of_backingstores ; number++ ))
    do 
        bucketclass_names+=("bucket.class$((number+1))")
        backingstore+=("compatible$((number+1))")
    done

    noobaa bucketclass create ${bucketclass_names[0]} --backingstores ${backingstore[0]}
    # noobaa bucketclass create ${bucketclass_names[1]} --placement Mirror --backingstores nb1,aws1 ❌ 
    # noobaa bucketclass create ${bucketclass_names[2]} --placement Spread --backingstores aws1,aws2 ❌ 
    noobaa bucketclass create ${bucketclass_names[3]} --backingstores ${backingstore[0]},${backingstore[1]}

    local bucketclass_list_array=($(noobaa silence bucketclass list | awk '{print $1}' | grep -v NAME))
    for bucketclass in ${bucketclass_list_array[@]}
    do
        noobaa bucketclass status ${bucketclass}
    done

    #TODO: activate the code below when we create all the bucketclass
    # if [ ${#bucketclass_list_array[@]} -ne $((${#bucketclass_names[@]}+1)) ]
    # then
    #     echo "❌  Bucket expected $((${#bucketclass_names[@]}+1)), and got ${#bucketclass_list_array[@]}."
    #     echo "👓  bucketclass list is ${bucketclass_list_array[@]}, Exiting."
    #     exit 1
    # fi

    noobaa status
    kuberun get bucketclass
    kuberun describe bucketclass
    echo "✅  bucketclass cycle is done"
}

function obc_cycle {
    echo "Starting the obc cycle"
    local bucket
    local buckets=()

    local bucketclass_list_array=($(noobaa silence bucketclass list | awk '{print $1}' | grep -v NAME | grep -v noobaa-default-bucket-class))
    for bucketclass in ${bucketclass_list_array[@]}
    do 
        buckets+=("bucket${bucketclass//[a-zA-Z.-]/}")
        if [ "${bucketclass//[a-zA-Z.-]/}" == "3" ]
        then
            flag="--app-namespace default"
        fi
        noobaa obc create ${buckets[$((${#buckets[@]}-1))]} --bucketclass ${bucketclass} ${flag}
        unset flag
    done
    noobaa obc list
    for bucket in ${buckets[@]}
    do
        noobaa obc status ${bucket}
    done
    kuberun get obc
    kuberun describe obc
    kuberun get obc,ob,secret,cm -l noobaa-obc

    # aws s3 --endpoint-url XXX ls
    echo "✅  obc cycle is done"
}

function delete_backingstore_path {
    local object_bucket backing_store
    local backingstore=($(noobaa silence backingstore list | grep -v "NAME" | awk '{print $1}'))
    local bucketclass=($(noobaa silence bucketclass list  | grep ${backingstore[0]} | awk '{print $1}'))
    local obc=($(noobaa silence obc list | grep ${backingstore[0]} | awk '{print $2}'))
    echo "Starting the delete related ${backingstore[0]} paths"
    if [ ${#obc[@]} -ne 0 ]
    then
        for object_bucket in ${obc[@]}
        do 
            noobaa obc delete ${object_bucket}
        done
    fi

    if [ ${#bucketclass[@]} -ne 0 ]
    then
        for bucket_class in ${bucketclass[@]}
        do 
            noobaa bucketclass delete ${bucket_class}
        done
    fi
    
    noobaa backingstore delete ${backingstore[0]}
    echo "✅  delete ${backingstore[0]} path is done"
}

function check_deletes {
    echo "Starting the delete cycle"
    local obc=($(noobaa silence obc list | grep -v "NAME\|default" | awk '{print $2}'))
    local bucketclass=($(noobaa silence bucketclass list  | grep -v NAME | awk '{print $1}'))
    local backingstore=($(noobaa silence backingstore list | grep -v "NAME" | awk '{print $1}'))
    noobaa obc delete ${obc[0]}
    noobaa bucketclass delete ${bucketclass[0]}
    noobaa backingstore list
    delete_backingstore_path
    echo "✅  delete cycle is done"
}
