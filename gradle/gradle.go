package gradle

import (
    "fmt"
    "github.com/bitrise-io/go-steputils/stepconf"
    "github.com/bitrise-io/go-utils/log"
    "github.com/bitrise-steplib/bitrise-step-build-orchestrator/execmd"
    "github.com/bitrise-steplib/bitrise-step-build-orchestrator/util"
)

type BuildConfig struct {
    Module      string     `env:"module,required"`
    DeployDir   string     `env:"deploy_path,required"`
    APK         string     `env:"target_apk,required"`
}

var gradlew = "./gradlew"

func Assemble() {
    var cfg BuildConfig
    if err := stepconf.Parse(&cfg); err != nil {
        util.Failf("Issue with an input: %s", err)
    }

    variant := "InternalDebugAndroidTest"

    log.Infof("Building %s %s", cfg.Module, variant)

     cmd := fmt.Sprintf("%s:assemble%s", cfg.Module, variant)
     execmd.ExecuteRelativeCommand(gradlew, cmd)
}

func PrepareForDeploy() {
    var cfg BuildConfig
    if err := stepconf.Parse(&cfg); err != nil {
        util.Failf("Issue with an input: %s", err)
    }
    execmd.ExecuteCommand("find", ".", "-name", cfg.APK, "-exec", "cp", "{}", cfg.DeployDir, ";")
}
