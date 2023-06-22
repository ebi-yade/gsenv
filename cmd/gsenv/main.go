package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/cockroachdb/errors"
	"github.com/hashicorp/logutils"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/iterator"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	logLevelCandidates := []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"}
	logFilter := &logutils.LevelFilter{
		Levels:   logLevelCandidates,
		MinLevel: logutils.LogLevel("INFO"),
		Writer:   os.Stderr,
	}
	logLevel := os.Getenv("GSENV_LOG_LEVEL")
	for _, candidate := range logLevelCandidates {
		if string(candidate) == logLevel {
			log.Println("set log level to", logLevel)
			logFilter.MinLevel = candidate
		}
	}
	log.SetOutput(logFilter)

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	var (
		flagProjectID string
		flagFilter    string
	)
	flag.StringVar(&flagProjectID, "project", "", "Google Cloud project id")
	flag.StringVar(&flagFilter, "filter", "", "Filter lists of secrets (according to the API spec)")
	flag.Parse()
	if flagProjectID != "" {
		projectID = flagProjectID
		if projectID == "" {
			return errors.New("--project is required but empty")
		}
	}
	log.Println("[DEBUG] projectID:", projectID)

	// fetch secrets from Google Cloud Secret Manager
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create secret manager client")
	}

	itr := client.ListSecrets(ctx, &secretmanagerpb.ListSecretsRequest{
		Parent: "projects/" + projectID,
		Filter: flagFilter,
	})

	eg := &errgroup.Group{}
	eg.SetLimit(25)
	for {
		secret, err := itr.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return errors.Wrap(err, "error Next")
		}
		log.Println("[DEBUG] ListSecretsResponse:", secret)
		secretName := secret.Name
		eg.Go(func() error {
			secretValue, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
				Name: secretName + "/versions/latest",
			})
			if err != nil {
				return errors.Wrap(err, "error AccessSecretVersion")
			}
			log.Println("[DEBUG] AccessSecretVersionResponse:", secretValue)
			if err := os.Setenv(basename(secretName), string(secretValue.GetPayload().GetData())); err != nil {
				return errors.Wrap(err, "error os.Setenv")
			}
			if err := os.Setenv(basename(secretName), string(secretValue.GetPayload().GetData())); err != nil {
				return errors.Wrap(err, "error os.Setenv")
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return errors.Wrap(err, "error eg.Wait")
	}

	execCommand := flag.Args()
	log.Println("[DEBUG] executing command:", execCommand)
	bin, err := exec.LookPath(execCommand[0])
	if err != nil {
		return err
	}
	if err := syscall.Exec(bin, execCommand, os.Environ()); err != nil {
		return errors.Wrap(err, "error syscall.Exec")
	}
	return nil
}

func basename(input string) string {
	aux := strings.Split(input, "/")
	return aux[len(aux)-1]
}
