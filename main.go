/*
Copyright 2023 Senjuti De

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

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	
	tektonClientset, err := versioned.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// Create Task
	taskClient := tektonClientset.TektonV1beta1().Tasks("default")

	task := &v1beta1.Task{
    ObjectMeta: metav1.ObjectMeta{
        Name: "sample-task",
    },
    Spec: v1beta1.TaskSpec{
        Steps: []v1beta1.Step{
            {
                Name: "step1",
                Image: "ubuntu",
                Command: []string{
                    "echo",
                    "Hello, Tekton!",
                },
            },
        },
    },
}


	fmt.Println("Creating Task...")
	resultTask, err := taskClient.Create(context.TODO(), task, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created Task %q.\n", resultTask.GetObjectMeta().GetName())
	fmt.Printf("Created Task %q. Image: %s, Message: %s\n", resultTask.GetObjectMeta().GetName(), resultTask.Spec.Steps[0].Image, resultTask.Spec.Steps[0].Command)

	// Update Task
	prompt()
	fmt.Println("Updating Task...")
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		resultTask, getErr := taskClient.Get(context.TODO(), "sample-task", metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("Failed to get the latest version of Task: %v", getErr))
		}

		// Print image and command before update
		fmt.Printf("Before Update - Image: %s, Command: %v\n", resultTask.Spec.Steps[0].Image, resultTask.Spec.Steps[0].Command)

		// Modify the Task as needed
		// For example, change the image to busybox and update the message
		resultTask.Spec.Steps[0].Image = "busybox"
		resultTask.Spec.Steps[0].Command = []string{"echo", "Updated Hello Tekton"}

		// Print image and command after update
		fmt.Printf("After Update - Image: %s, Command: %v\n", resultTask.Spec.Steps[0].Image, resultTask.Spec.Steps[0].Command)

		_, updateErr := taskClient.Update(context.TODO(), resultTask, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("Update failed: %v", retryErr))
	}
	fmt.Println("Updated Task...")
	fmt.Printf("Updated Task %q. Image: %s, Message: %s\n", resultTask.GetObjectMeta().GetName(), resultTask.Spec.Steps[0].Image, resultTask.Spec.Steps[0].Command)

	// List Tasks
	prompt()
	fmt.Printf("Listing Tasks in namespace %q:\n", "default")
	taskList, err := taskClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, t := range taskList.Items {
		fmt.Printf(" * %s\n", t.Name)
	}

	// Delete Task
	prompt()
	fmt.Println("Deleting Task...")
	deletePolicy := metav1.DeletePropagationForeground
	if err := taskClient.Delete(context.TODO(), "sample-task", metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
	}
	fmt.Println("Deleted Task.")
}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}
