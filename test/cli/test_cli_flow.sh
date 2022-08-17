#!/bin/bash

export PS4='\e[36m+ ${FUNCNAME:-main}\e[0m@\e[32m${BASH_SOURCE}:\e[35m${LINENO} \e[0m'

#In first stage, the script assume that the noobaa cli is installed.
#Also assuming aws cli is installed

NAMESPACE='test'
CM=false

function post_install_tests {
    aws_credentials
    check_pv_pool_resources
    check_S3_compatible
    check_namespacestore
    create_replication_files
    bucketclass_cycle
    bz_2038884
    obc_cycle
    replication_cycle
    check_backingstore
    # check_dbdump
    account_cycle
    check_deletes
    delete_replication_files
    check_pgdb_config_override
    test_noobaa_cr_deletion
    test_noobaa_loadbalancer_source_subnet
}

function main {
    noobaa_install
    if [ "${CM}" == "true" ]
    then
        check_core_config_map
    else
        post_install_tests
    fi
    noobaa_uninstall
 }

function usage {
    set +x
    echo -e "\nUsage: ${0} [options]"
    echo "--namespace             -   Change the namespace (default: ${NAMESPACE})"
    echo "--mongo-image           -   Change the mongo image"
    echo "--noobaa-image          -   Change the noobaa image"
    echo "--operator-image        -   Change the operator image"
    echo "--check_core_config_map -   Check Only core config map"
    echo -e "--help               -   print this help\n"
    exit 1
}

function set_nonstandard_options {
    if [ ! -z ${MONGO_IMAGE} ]
    then
        noobaa+=" --mongo-image ${MONGO_IMAGE}"
    fi

    if [ ! -z ${NOOBAA_IMAGE} ]
    then
        noobaa+=" --noobaa-image ${NOOBAA_IMAGE}"
    fi

    if [ ! -z ${OPERATOR_IMAGE} ]
    then
        noobaa+=" --operator-image ${OPERATOR_IMAGE}"
    fi
}

while true
do
    if [ -z ${1} ]; then
        break
    fi

    case ${1} in
        --mongo-image)           MONGO_IMAGE=${2}
                                 shift 2;;
        --noobaa-image)          NOOBAA_IMAGE=${2}
                                 shift 2;;
        --operator-image)        OPERATOR_IMAGE=${2}
                                 shift 2;;
        -n|--namespace)          NAMESPACE=${2}
                                 shift 2;;
        --check_core_config_map) CM=true
                                 shift;;       
        -h|--help)               usage;;
        *)                       usage;;
    esac
done

#Setting noobaa command with namespace
#The reason that we are doing it in a variable and not alias is
#That alias is not expended in non interactive shell
#Currently will work only on noobaa-operator-local - need to change it
noobaa="build/_output/bin/noobaa-operator-local -n ${NAMESPACE}"
kubectl="kubectl -n ${NAMESPACE}"
. $(dirname ${0})/test_cli_functions.sh

#Setting the noobaa command with non standard options if needed.
set_nonstandard_options

main
