package main

import (
    "fmt"
    "os"
    "strings"
    "time"
    "github.com/bitrise-io/go-steputils/stepconf"
    "github.com/bitrise-io/go-utils/log"
    "github.com/bitrise-steplib/bitrise-step-build-orchestrator/deploy"
    "github.com/bitrise-steplib/bitrise-step-build-orchestrator/env"
    "github.com/bitrise-steplib/bitrise-step-build-orchestrator/execmd"
    "github.com/bitrise-steplib/bitrise-step-build-orchestrator/gh"
    "github.com/bitrise-steplib/bitrise-step-build-orchestrator/gradle"
    "github.com/bitrise-steplib/bitrise-step-build-orchestrator/trigger"
    "github.com/bitrise-steplib/bitrise-step-build-orchestrator/util"
)

func DisplayInfo() {
    log.Infof("=== Display environment info ===")
    execmd.ExecuteCommand("go", "version")
    execmd.ExecuteCommand("git", "--version")
    execmd.ExecuteCommand("adb", "--version")
    execmd.ExecuteRelativeCommand("./gradlew", "--version")
}

type PathConfig struct {
    Type        string     `env:"module_type,opt[feature,root,testing,design]"`
    Forced      bool       `env:"forced"`
    Modules     string     `env:"modules_to_test_list"`
}

func checkIfTestsExist(testPath string) bool {
    log.Infof("Checking for tests in %s", testPath)
    var root = fmt.Sprintf("./%s", testPath)
    if _, err := os.Stat(root); !os.IsNotExist(err) {
        log.Infof("OK. Found %s!\n", testPath)
        return true
    }
    return false
}

func isSkippable(module string) bool {

    var cfg PathConfig
    if err := stepconf.Parse(&cfg); err != nil {
        util.Failf("Issue with an input: %s", err)
    }

    if cfg.Forced {
        log.Infof("Forcing build")
        return false
    }

    if _, detected := os.LookupEnv("TARGET_MODULE"); detected {
        log.Infof("TARGET_MODULE detected, forcing build")
        return false
    }

    var testModuleDir string
    var testPath string
    var targetModule string

    switch cfg.Type {
    case "root":
        testModuleDir = module
        testPath = fmt.Sprintf("%s/src/androidTest", testModuleDir)
        targetModule = module
    case "testing":
        testModuleDir = strings.TrimSuffix(module, "-tests")
        testPath = fmt.Sprintf("testing/%s/src/androidTest", testModuleDir)
        targetModule = "testing"
    case "design":
        testModuleDir = strings.TrimSuffix(strings.TrimPrefix(module, "design-"), "-tests")
        testPath = fmt.Sprintf("design/%s/src/androidTest", testModuleDir)
        targetModule = "design"
    default:
        testModuleDir = strings.TrimPrefix(module, "feature-")
        testPath = fmt.Sprintf("features/%s/src/androidTest", testModuleDir)
        targetModule = module
    }

    exists := checkIfTestsExist(testPath)
    if !exists {
        log.Errorf("No tests detected in %s. Skipping build", module)
        return true
    }

    modules := gh.GetChangedModules()
    if modules[targetModule] == false {
        log.Errorf("No changes detected in %s. Skipping build", targetModule)
        return true
    }
    log.Infof("Changes to module %s found. Running tests.", targetModule)
    return false
}

func buildAndTrigger() {
    timestamp()
    gradle.Assemble()
    timestamp()
    gradle.PrepareForDeploy()
    timestamp()
    env.SetTargetEnv()
    timestamp()
    deploy.Deploy()
    timestamp()
    trigger.TriggerWorkflow()
    timestamp()
}

var startTime int64 = 0

func timestamp() {
    log.Infof("[Time] %d", time.Now().UnixNano() / int64(time.Millisecond) - startTime)
}

func camelCase(s []string) string {
    var g []string
    for _,value := range s {
        g = append(g,strings.Title(value))
    }
    return strings.Join(g,"")
}

func exportModuleConfiguration(module string, index int) {

    // Module:
    // feature-account-closure

    // Variant:
    // InternalDebugAndroidTest

    // Target APK:
    // feature-account-closure-internal-debug-androidTest.apk

    // Test Package:
    // com.neofinancial.neo.account.closure.test

    // Test Runner:
    // com.neofinancial.neo.testing.AccountClosureAndroidTestRunner

    // Workflows:
    // test-worker-account-closure

    moduleComponents := strings.Split(module, "-")[1:]
    targetApk := module + "-internal-debug-androidTest.apk"
    testPackage := "com.neofinancial.neo." + strings.Join(moduleComponents, ".") + ".test"
    testRunner := "com.neofinancial.neo.testing." + camelCase(moduleComponents) + "AndroidTestRunner"
    workflow := fmt.Sprintf("test-worker-agent-%02d", index)

    log.Infof("Set environment target_apk variable to [%s]", targetApk)
    log.Infof("Set environment test_package variable to [%s]", testPackage)
    log.Infof("Set environment test_runner variable to [%s]", testRunner)
    log.Infof("Set environment module variable to [%s]", module)
    log.Infof("Set environment workflows variable to [%s]", workflow)

    os.Setenv("target_apk", targetApk)
    os.Setenv("test_package", testPackage)
    os.Setenv("test_runner", testRunner)
    os.Setenv("module", module)
    os.Setenv("workflows", workflow)

    // Environment to share:
    // PARENT_SLUG
    // TARGET_APK
    // ADB_COMMAND
    // MODULE_NAME
}

func main() {
    startTime = time.Now().UnixNano() / int64(time.Millisecond)
    timestamp()
    var cfg PathConfig
    if err := stepconf.Parse(&cfg); err != nil {
        util.Failf("Issue with an input: %s", err)
    }
    timestamp()
    DisplayInfo()

    moduleList := strings.Split(cfg.Modules, "\n")
    for i, module := range moduleList {
        if module == "" {
            continue
        }
        log.Infof("Request to build module %d %s", i, module)

        if isSkippable(module) {
            os.Exit(0)
        }

        exportModuleConfiguration(module, i)

        log.Infof("Building %s", module)
        timestamp()
        buildAndTrigger()
    }

    os.Exit(0)
}