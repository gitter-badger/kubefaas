/*
Copyright 2019 The Fission Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mqtrigger

import (
	"fmt"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	fv1 "github.com/srcmesh/kubefaas/pkg/apis/core/v1"
	"github.com/srcmesh/kubefaas/pkg/cli/cliwrapper/cli"
	"github.com/srcmesh/kubefaas/pkg/cli/cmd"
	"github.com/srcmesh/kubefaas/pkg/cli/cmd/spec"
	"github.com/srcmesh/kubefaas/pkg/cli/console"
	flagkey "github.com/srcmesh/kubefaas/pkg/cli/flag/key"
	"github.com/srcmesh/kubefaas/pkg/cli/util"
	"github.com/srcmesh/kubefaas/pkg/mqtrigger/validator"
)

type CreateSubCommand struct {
	cmd.CommandActioner
	trigger *fv1.MessageQueueTrigger
}

func Create(input cli.Input) error {
	return (&CreateSubCommand{}).do(input)
}

func (opts *CreateSubCommand) do(input cli.Input) error {
	err := opts.complete(input)
	if err != nil {
		return err
	}
	return opts.run(input)
}

func (opts *CreateSubCommand) complete(input cli.Input) error {
	mqtName := input.String(flagkey.MqtName)
	if len(mqtName) == 0 {
		console.Warn(fmt.Sprintf("--%v will be soon marked as required flag, see 'help' for details", flagkey.MqtName))
		mqtName = uuid.NewV4().String()
	}
	fnName := input.String(flagkey.MqtFnName)
	fnNamespace := input.String(flagkey.NamespaceFunction)

	mqType := (fv1.MessageQueueType)(input.String(flagkey.MqtMQType))
	if !validator.IsValidMessageQueue((string)(mqType)) {
		return errors.New("Unsupported message queue type")
	}

	topic := input.String(flagkey.MqtTopic)
	if len(topic) == 0 {
		return errors.New("topic cannot be empty")
	}

	respTopic := input.String(flagkey.MqtRespTopic)
	if topic == respTopic {
		// TODO maybe this should just be a warning, perhaps
		// allow it behind a --force flag
		return errors.New("listen topic should not equal to response topic")
	}

	errorTopic := input.String(flagkey.MqtErrorTopic)
	maxRetries := input.Int(flagkey.MqtMaxRetries)

	if maxRetries < 0 {
		return errors.New("Maximum number of retries must be greater than or equal to 0")
	}

	contentType := input.String(flagkey.MqtMsgContentType)
	if len(contentType) == 0 {
		contentType = "application/json"
	}

	err := checkMQTopicAvailability(mqType, topic, respTopic)
	if err != nil {
		return err
	}

	if input.Bool(flagkey.SpecSave) {
		specDir := util.GetSpecDir(input)
		fr, err := spec.ReadSpecs(specDir)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error reading spec in '%v'", specDir))
		}

		exists, err := fr.ExistsInSpecs(fv1.Function{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fnName,
				Namespace: fnNamespace,
			},
		})
		if err != nil {
			return err
		}
		if !exists {
			console.Warn(fmt.Sprintf("MessageQueueTrigger '%v' references unknown Function '%v', please create it before applying spec",
				mqtName, fnName))
		}
	}

	opts.trigger = &fv1.MessageQueueTrigger{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mqtName,
			Namespace: fnNamespace,
		},
		Spec: fv1.MessageQueueTriggerSpec{
			FunctionReference: fv1.FunctionReference{
				Type: fv1.FunctionReferenceTypeFunctionName,
				Name: fnName,
			},
			MessageQueueType: mqType,
			Topic:            topic,
			ResponseTopic:    respTopic,
			ErrorTopic:       errorTopic,
			MaxRetries:       maxRetries,
			ContentType:      contentType,
		},
	}

	return nil
}

func (opts *CreateSubCommand) run(input cli.Input) error {
	// if we're writing a spec, don't call the API
	// save to spec file or display the spec to console
	if input.Bool(flagkey.SpecDry) {
		return spec.SpecDry(*opts.trigger)
	}

	if input.Bool(flagkey.SpecSave) {
		specFile := fmt.Sprintf("mqtrigger-%v.yaml", opts.trigger.ObjectMeta.Name)
		err := spec.SpecSave(*opts.trigger, specFile)
		if err != nil {
			return errors.Wrap(err, "error saving message queue trigger spec")
		}
		return nil
	}

	_, err := opts.Client().V1().MessageQueueTrigger().Create(opts.trigger)
	if err != nil {
		return errors.Wrap(err, "create message queue trigger")
	}

	fmt.Printf("trigger '%s' created\n", opts.trigger.ObjectMeta.Name)
	return nil
}

func checkMQTopicAvailability(mqType fv1.MessageQueueType, topics ...string) error {
	for _, t := range topics {
		if len(t) > 0 && !validator.IsValidTopic((string)(mqType), t) {
			return errors.Errorf("invalid topic for %s: %s", mqType, t)
		}
	}
	return nil
}