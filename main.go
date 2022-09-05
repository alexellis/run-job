package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/google/uuid"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"

	"k8s.io/client-go/rest"

	"k8s.io/client-go/tools/clientcmd"
)

var (
	Version   string
	GitCommit string
)

func main() {

	var (
		kubeconfig string
		namespace  string
		name       string
		sa         string
		image      string
		outFile    string
	)

	flag.StringVar(&namespace, "namespace", "default", "Namespace for the job")
	flag.StringVar(&name, "name", "job1", "Name of the job")
	flag.StringVar(&sa, "sa", "", "Service account to use for the job")
	flag.StringVar(&image, "image", "", "Image for the job")
	flag.StringVar(&outFile, "out", "", "File to write to or leave blank for STDOUT")

	flag.StringVar(&kubeconfig, "kubeconfig", "$HOME/.kube/config", "Path to KUBECONFIG")
	flag.Parse()

	if len(name) == 0 {
		log.Fatalf("--job is required")
	}
	if len(image) == 0 {
		log.Fatalf("--image is required")
	}

	clientset, err := getClientset(kubeconfig)
	if err != nil {
		panic(err)
	}

	u := uuid.New()

	parallelism := int32(1)
	ctx := context.Background()
	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app":    "run-job",
				"job-id": u.String(),
			},
		},
		Spec: batchv1.JobSpec{
			Parallelism:  &parallelism,
			BackoffLimit: &parallelism,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":    "run-job",
						"job-id": u.String(),
					},
					Name:      name,
					Namespace: namespace,
				},
				Spec: corev1.PodSpec{
					RestartPolicy:      corev1.RestartPolicyNever,
					ServiceAccountName: sa,
					Containers: []corev1.Container{
						{
							Image:           image,
							Name:            name,
							ImagePullPolicy: corev1.PullAlways,
						},
					},
				},
			},
		},
	}

	job, err := clientset.BatchV1().Jobs(namespace).Create(ctx, jobSpec, metav1.CreateOptions{})

	if err != nil {
		log.Fatalf("Error creating job: %v", err)
	}

	fmt.Printf("Created job %q (%s)\n", job.GetName(), u.String())

	if err := watchForJob(clientset, ctx, u.String(), job.GetName(), job.Namespace, outFile); err != nil {
		log.Fatalf("Error watching job: %v", err)
	}
}

func watchForJob(clientset *kubernetes.Clientset, ctx context.Context, jobID, name, namespace, outFile string) error {

	listOptions := metav1.ListOptions{
		LabelSelector: "job-id=" + jobID,
	}

	wCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	rw, err := watchtools.NewRetryWatcher("1", &cache.ListWatch{
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return clientset.BatchV1().Jobs(namespace).Watch(wCtx, listOptions)
		},
	})
	if err != nil {
		return fmt.Errorf("error creating label watcher: %s", err.Error())
	}

	go func() {
		<-ctx.Done()
		// Cancel the context
		rw.Stop()
	}()

	ch := rw.ResultChan()
	//defer rw.Stop()

	for event := range ch {
		// We need to inspect the event and get ResourceVersion out of it
		switch event.Type {

		case watch.Added, watch.Modified:
			job, ok := event.Object.(*batchv1.Job)
			if !ok {
				return fmt.Errorf("unable to parse Kubernetes Job from Annotation watcher")
			}

			done := false
			message := ""
			failed := false

			for _, condition := range job.Status.Conditions {
				switch condition.Type {
				case batchv1.JobFailed:
					failed = true
					message = condition.Message

					done = true
				case batchv1.JobComplete:
					failed = false
					message = condition.Message

					done = true
				}
			}

			fmt.Printf(".")

			if done {
				if failed {
					fmt.Printf("\nJob %s (%s) failed %s\n", name, jobID, message)
				} else {
					fmt.Printf("\nJob %s (%s) succeeded %s\n", name, jobID, message)
				}

				pods, err := clientset.CoreV1().Pods(namespace).List(wCtx, listOptions)
				if err == nil {
					podNames := []string{}
					for _, pod := range pods.Items {
						podNames = append(podNames, pod.Name)
					}

					logsOut, err := logs(wCtx, clientset, podNames, namespace)
					if err != nil {
						log.Fatalf("Error getting logs: %v", err)
					}
					logsOut = "Recorded: " + time.Now().UTC().String() + "\n\n" + logsOut

					if len(outFile) > 0 {
						if err := ioutil.WriteFile(outFile, []byte(logsOut), os.ModePerm); err != nil {
							log.Fatalf("Error writing logs to file: %v", err)
						} else {
							log.Println("Logs written to file", outFile)
						}
					} else {
						fmt.Printf("\n%s\n", logsOut)
					}
					deletePol := metav1.DeletePropagationBackground
					if err := clientset.BatchV1().Jobs(namespace).Delete(wCtx, name, metav1.DeleteOptions{PropagationPolicy: &deletePol}); err != nil {
						log.Fatalf("Error deleting job: %v", err)
					} else {
						fmt.Printf("Deleted job %s\n", name)
					}

				} else {
					log.Fatalf("Error getting pods: %v", err)
				}

				cancel()

				return nil
			}

		case watch.Deleted:
			_, ok := event.Object.(*batchv1.Job)
			if !ok {
				return fmt.Errorf("unable to parse Kubernetes Job from Annotation watcher")
			}

		case watch.Bookmark:
			// Un-used
		case watch.Error:
			log.Printf("Error attempting to watch Kubernetes Jobs")

			// This round trip allows us to handle unstructured status
			errObject := apierrors.FromObject(event.Object)
			statusErr, ok := errObject.(*apierrors.StatusError)
			if !ok {
				log.Printf(spew.Sprintf("received an error which is not *metav1.Status but %#+v", event.Object))

			}

			status := statusErr.ErrStatus
			log.Printf("%v", status)
		default:
		}
	}

	return nil

}

func getClientset(kubeconfig string) (*kubernetes.Clientset, error) {

	kubeconfig = strings.ReplaceAll(kubeconfig, "$HOME", os.Getenv("HOME"))
	kubeconfig = strings.ReplaceAll(kubeconfig, "~", os.Getenv("HOME"))
	masterURL := ""

	var clientConfig *rest.Config
	if _, err := os.Stat(kubeconfig); err != nil {
		config, err := rest.InClusterConfig()
		if err != nil {
			log.Fatalf("Error building in-cluster config: %s", err.Error())
		}
		clientConfig = config
	} else {
		config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
		if err != nil {
			log.Fatalf("Error building kubeconfig: %s %s", kubeconfig, err.Error())
		}
		clientConfig = config
	}

	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func logs(ctx context.Context, clientset *kubernetes.Clientset, pods []string, namespace string) (string, error) {
	buf := new(bytes.Buffer)

	for _, pod := range pods {
		req := clientset.CoreV1().Pods(namespace).GetLogs(pod, &corev1.PodLogOptions{})
		stream, err := req.Stream(ctx)
		if err != nil {
			return "", fmt.Errorf("error while reading %s logs %w", pod, err)
		}

		_, err = io.Copy(buf, stream)
		stream.Close()
		if err != nil {
			return "", fmt.Errorf("error while reading %s logs %w", pod, err)
		}
	}

	return buf.String(), nil
}
