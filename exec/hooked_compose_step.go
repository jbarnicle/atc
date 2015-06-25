package exec

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/tedsuo/ifrit"
)

type hookedCompose struct {
	step    StepFactory
	failure StepFactory
	success StepFactory
	ensure  StepFactory
	next    StepFactory

	prev Step
	repo *SourceRepository

	firstStep   Step
	failureStep Step
	successStep Step
	ensureStep  Step
	nextStep    Step
}

func HookedCompose(
	step StepFactory,
	next StepFactory,
	failure StepFactory,
	success StepFactory,
	ensure StepFactory,
) StepFactory {
	return hookedCompose{
		step:    step,
		next:    next,
		failure: failure,
		success: success,
		ensure:  ensure,
	}
}

func (hc hookedCompose) Using(prev Step, repo *SourceRepository) Step {
	hc.repo = repo
	hc.prev = prev
	return &hc
}

func (hc *hookedCompose) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	hc.firstStep = hc.step.Using(hc.prev, hc.repo)

	firstStepError := hc.backgroundProcess(hc.firstStep, signals)

	var succeeded Success

	// if whatever step I just ran cannot respond to success, we want to return a noop
	if !hc.firstStep.Result(&succeeded) {
		noop := &NoopStep{}
		return noop.Run(signals, ready)
	}

	hc.ensureStep = hc.ensure.Using(hc.firstStep, hc.repo)
	hooks := []Step{hc.ensureStep}

	if firstStepError == nil {
		if bool(succeeded) {
			hc.successStep = hc.success.Using(hc.firstStep, hc.repo)
			hooks = append(hooks, hc.successStep)
		} else {
			hc.failureStep = hc.failure.Using(hc.firstStep, hc.repo)
			hooks = append(hooks, hc.failureStep)
		}
	}

	var hookError chan error
	hookError = make(chan error)

	for _, hook := range hooks {
		go func(hookProcess ifrit.Runner) {
			hookError <- hc.backgroundProcess(hookProcess, signals)
		}(hook)
	}

	var errors error

	var err error

	for i := 0; i < len(hooks); i++ {
		err = <-hookError

		if err != nil {
			errors = multierror.Append(errors, err)
		}
	}

	var allHooksSuccessful Success
	allHooksSuccessful = true
	for _, hook := range hooks {
		var hookSuccessful Success

		if !hook.Result(&hookSuccessful) {
			allHooksSuccessful = false
		}

		if !bool(hookSuccessful) {
			allHooksSuccessful = false
		}
	}

	if firstStepError != nil {
		errors = multierror.Append(errors, firstStepError)
	}

	if errors != nil {
		return errors
	}

	if bool(succeeded) && bool(allHooksSuccessful) {
		hc.nextStep = hc.next.Using(hc.firstStep, hc.repo)
	} else {
		hc.nextStep = &NoopStep{}
	}

	return hc.nextStep.Run(signals, ready)
}

func (hc *hookedCompose) backgroundProcess(stepProcess ifrit.Runner, signals <-chan os.Signal) error {
	var signalled bool
	var waitErr error

	process := ifrit.Background(stepProcess)

dance:
	for {
		select {
		case waitErr = <-process.Wait():
			break dance

		case sig := <-signals:
			process.Signal(sig)
			signalled = true
		}
	}

	if signalled || waitErr != nil {
		return waitErr
	}

	return nil
}

func (hc *hookedCompose) Release() error {
	errorMessages := []string{}
	if hc.firstStep != nil {
		if err := hc.firstStep.Release(); err != nil {
			errorMessages = append(errorMessages, "first step: "+err.Error())
		}
	}

	if hc.ensureStep != nil {
		if err := hc.ensureStep.Release(); err != nil {
			errorMessages = append(errorMessages, "ensure step: "+err.Error())
		}
	}

	if hc.failureStep != nil {
		if err := hc.failureStep.Release(); err != nil {
			errorMessages = append(errorMessages, "failure step: "+err.Error())
		}
	}

	if hc.successStep != nil {
		if err := hc.successStep.Release(); err != nil {
			errorMessages = append(errorMessages, "success step: "+err.Error())
		}
	}

	if hc.nextStep != nil {
		if err := hc.nextStep.Release(); err != nil {
			errorMessages = append(errorMessages, "next step: "+err.Error())
		}
	}

	if len(errorMessages) > 0 {
		return fmt.Errorf("sources failed to release:\n%s", strings.Join(errorMessages, "\n"))
	}

	return nil
}

func (hc *hookedCompose) Result(x interface{}) bool {
	return hc.nextStep.Result(x)
}
