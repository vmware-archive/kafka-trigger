/*
Copyright (c) 2016-2017 Bitnami

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

package cronjob

import (
	"fmt"
	"io"

	"github.com/gosuri/uitable"
	"github.com/kubeless/cronjob-trigger/pkg/client/clientset/versioned"
	cronjobUtils "github.com/kubeless/cronjob-trigger/pkg/utils"
	kubelessUtils "github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var listCmd = &cobra.Command{
	Use:     "list FLAG",
	Aliases: []string{"ls"},
	Short:   "list all Cronjob triggers deployed to Kubeless",
	Long:    `list all Cronjob triggers deployed to Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err.Error())
		}
		if ns == "" {
			ns = kubelessUtils.GetDefaultNamespace()
		}

		kubelessClient, err := cronjobUtils.GetKubelessClientOutCluster()
		if err != nil {
			logrus.Fatalf("Can not create out-of-cluster client: %v", err)
		}

		if err := doList(cmd.OutOrStdout(), kubelessClient, ns); err != nil {
			logrus.Fatal(err.Error())
		}
	},
}

func init() {
	listCmd.Flags().StringP("namespace", "n", "", "Specify namespace for the function")
}

func doList(w io.Writer, kubelessClient versioned.Interface, ns string) error {
	triggersList, err := kubelessClient.KubelessV1beta1().CronJobTriggers(ns).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	table := uitable.New()
	table.MaxColWidth = 50
	table.Wrap = true
	table.AddRow("NAME", "NAMESPACE", "SCHEDULE", "FUNCTION NAME")
	for _, trigger := range triggersList.Items {
		table.AddRow(trigger.Name, trigger.Namespace, trigger.Spec.Schedule, trigger.Spec.FunctionName)
	}
	fmt.Fprintln(w, table)
	return nil
}
