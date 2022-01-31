#!/bin/bash

NOT_CHANGED=($(./gradlew -q printModulesFromPaths -PprintAll=true))
ALL_MODULES=( $(echo "${NOT_CHANGED[1]}" | sed 's/NOT_CHANGED=//g' | sed 's/\,/ /g') )

for MODULE in "${ALL_MODULES[@]}"
do
  TEST_PATH=$(echo "$MODULE/src/androidTest" | sed 's/feature\-/\.\/features\//g')
  TEST_WORKER=$(echo "test-worker-${MODULE}" | sed 's/feature\-//g')
  if [[ -z "${FORCE_ORCHESTRATOR}" ]]; then
    envman add --key "skip-${TEST_WORKER}" --value true
  else
    envman add --key "skip-${TEST_WORKER}" --value false
  fi
done

if [[ -z "${PULL_REQUEST_ID}" ]]; then
  echo "Not a pull request, skipping"
  exit 0
fi

if [[ ! -z "${FORCE_ORCHESTRATOR}" ]]; then
  echo "Force orchestrator"
  exit 0
fi

PR_NUMBER=$PULL_REQUEST_ID

echo "Checking $PR_NUMBER"

FILES=( $(curl -s -X GET -H "Accept: application/vnd.github.v3+json" -H "Authorization: token $GITHUB_TOKEN" "https://api.github.com/repos/neofinancial/neo-android/pulls/$PR_NUMBER/files" | jq ".[$i] .filename") )

MODULE_PATHS=()

for FILE in "${FILES[@]}"
do
  MODULE_PATH=`echo $FILE | sed -n 's/^\"\(.*\)\/src\/.*$/\1/p'`
  MODULE_PATHS+=($MODULE_PATH)
done

MODULE_PATHS=($(printf "%s\n" "${MODULE_PATHS[@]}" | sort -u))
MODULE_PATH_STRING=${MODULE_PATHS[@]}
MODULE_PATH_STRING=${MODULE_PATH_STRING// /,}

CHANGED_MODULES=($(./gradlew -q printModulesFromPaths -Ppaths=$MODULE_PATH_STRING -PprintAll=false | sed 's/\,/ /g'))

for MODULE in ${CHANGED_MODULES[@]}; do
  TEST_PATH=$(echo "$MODULE/src/androidTest" | sed 's/feature\-/\.\/features\//g')
  TEST_WORKER=$(echo "test-worker-${MODULE}" | sed 's/feature\-//g')

  if [ -d "$TEST_PATH" ]; then
    echo "Building and running tests in $TEST_WORKER"
    envman add --key "skip-${TEST_WORKER}" --value false
  else
    echo "Changes detected, but no tests found in $MODULE"
    envman add --key "skip-${TEST_WORKER}" --value true
  fi
done

if [ ! -z "$TARGET_MODULE" ]; then
  TEST_WORKER=$(echo skip-test-worker-$TARGET_MODULE | sed s/feature\-//g)
  echo "Force build of $TARGET_MODULE; Set $TEST_WORKER to false"
  envman add --key $TEST_WORKER --value false
fi




# Read all the environment variables
# Build the appropriate trigger query
# Trigger

# Root action command:
# bash ./scripts/bitrise-execute.sh $BITRISE_TOKEN $PULL_ID ${{ steps.workflow.outputs.workflow }} ${{ steps.module.outputs.module }}

        # inputs:
        # - workflows: test-worker-account-closure
        # - module: feature-account-closure
        # - variant: InternalDebugAndroidTest
        # - target_apk: feature-account-closure-internal-debug-androidTest.apk
        # - test_package: com.neofinancial.neo.account.closure.test
        # - test_runner: com.neofinancial.neo.testing.AccountClosureAndroidTestRunner
        # - environment_key_list: |-
        #     PARENT_SLUG
        #     TARGET_APK
        #     ADB_COMMAND
        #     MODULE_NAME
        # - access_token: "$BITRISE_KEY"



# BITRISE_TOKEN=$1
# PR_NUMBER=$2
# WORKFLOW=$3
# MODULE=$4

# API=`gh api repos/:owner/:repo/pulls/$PR_NUMBER`

# BRANCH=`echo $API | tr '\r\n' ' ' | jq .base.ref`
# BRANCH_DEST=`echo $API | tr '\r\n' ' ' | jq .head.ref`
# PULL_ID=$PR_NUMBER
# COMMIT_HASH=`echo $API | tr '\r\n' ' ' | jq .head.sha`

# if [ "$WORKFLOW" == "test" ] && [ ! -z "$MODULE" ]; then
#     WORKFLOW="build-requester"
# fi

# echo "Trigger $WORKFLOW on $PR_NUMBER for $MODULE"

# BUILD_URL=`curl -s -X POST -H "Authorization: $BITRISE_TOKEN" "https://api.bitrise.io/v0.1/apps/60ded9f99834a343/builds" -d \
# '{
#     "hook_info": {
#         "type":"bitrise"
#     },
#     "build_params": {
#         "workflow_id":"'$WORKFLOW'",
#         "branch":'$BRANCH',
#         "branch_dest":'$BRANCH_DEST',
#         "pull_request_id":'$PULL_ID',
#         "commit_hash":'$COMMIT_HASH',
#         "environments": [
#             {
#                 "is_expand": true,
#                 "mapped_to":"TARGET_MODULE",
#                 "value":"'$MODULE'"
#             }
#         ]
#     },
#     "triggered_by":"curl"
#   }' | jq -r .build_url`

# gh pr comment $PR_NUMBER --body "Bitrise is running workflow _${WORKFLOW}_: [Build link]($BUILD_URL)"

