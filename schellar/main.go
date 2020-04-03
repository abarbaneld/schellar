package main

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
)

var (
	conductorURL         string
	mongoURL             string
	mongoAddress         string
	mongoUsername        string
	mongoPassword        string
	dbName               = "admin"
	checkIntervalSeconds = 10
)

//Schedule struct data
type Schedule struct {
	Name                string                 `json:"name,omitempty" bson:"name"`
	Enabled             bool                   `json:"enabled,omitempty" bson:"enabled"`
	Status              string                 `json:"status,omitempty" bson:"status"`
	WorkflowName        string                 `json:"workflowName,omitempty" bson:"workflowName"`
	WorkflowVersion     string                 `json:"workflowVersion,omitempty" bson:"workflowVersion"`
	WorkflowContext     map[string]interface{} `json:"workflowContext,omitempty" bson:"workflowContext"`
	CronString          string                 `json:"cronString,omitempty" bson:"cronString"`
	ParallelRuns        bool                   `json:"parallelRuns,omitempty" bson:"parallelRuns"`
	CheckWarningSeconds int                    `json:"checkWarningSeconds,omitempty" bson:"checkWarningSeconds"`
	FromDate            *time.Time             `json:"fromDate,omitempty" bson:"fromDate"`
	ToDate              *time.Time             `json:"toDate,omitempty" bson:"toDate"`
	LastUpdate          time.Time              `json:"lastUpdate,omitempty" bson:"lastUpdate"`
}

func main() {
	logLevel := flag.String("loglevel", "debug", "debug, info, warning, error")
	checkInterval0 := flag.Int("check-interval", 10, "Workflow check interval in seconds")
	conductorURL0 := flag.String("conductor-api-url", "", "Conductor API URL. Example: http://conductor-server:8080/api")
	mongoURL0 := flag.String("mongo-url", "", "MongoDB Url. Example: 'mongdb://mongo1:1234/db1,mongo2:1234/db1")
	mongoAddress0 := flag.String("mongo-address", "", "MongoDB address. Example: 'mongo'")
	mongoUsername0 := flag.String("mongo-username", "root", "MongoDB username")
	mongoPassword0 := flag.String("mongo-password", "root", "MongoDB password")
	flag.Parse()

	switch *logLevel {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
		break
	case "warning":
		logrus.SetLevel(logrus.WarnLevel)
		break
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
		break
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	conductorURL = *conductorURL0
	if conductorURL == "" {
		logrus.Errorf("'conductor-api-url' parameter is required")
		os.Exit(1)
	}

	mongoURL = *mongoURL0
	mongoAddress = *mongoAddress0
	if mongoAddress == "" && mongoURL == "" {
		logrus.Errorf("'mongo-address' or 'mongo-url' parameter is required")
		os.Exit(1)
	}

	mongoUsername = *mongoUsername0
	mongoPassword = *mongoPassword0

	checkIntervalSeconds = *checkInterval0

	logrus.Info("====Starting Schellar====")

	logrus.Debugf("Connecting to MongoDB")
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    strings.Split(mongoAddress, ","),
		Timeout:  2 * time.Second,
		Database: dbName,
		Username: mongoUsername,
		Password: mongoPassword,
	}

	mongoDBDialInfo, er := mgo.ParseURL(mongoURL)
	if er != nil {
		logrus.Errorf("Error parsing mongdb URL. err=%s", er)
	}

	for i := 0; i < 30; i++ {
		ms, err := mgo.DialWithInfo(mongoDBDialInfo)

		if err != nil {
			logrus.Infof("Couldn't connect to mongdb. err=%s", err)
			time.Sleep(1 * time.Second)
			logrus.Infof("Retrying...")
			continue
		}
		mongoSession = ms
		logrus.Infof("Connected to MongoDB successfully")
		break
	}

	if mongoSession == nil {
		logrus.Errorf("Couldn't connect to MongoDB")
		os.Exit(1)
	}

	err := startScheduler()
	if err != nil {
		logrus.Errorf("Error during scheduler startup. err=%s", err)
		os.Exit(1)
	}
	startRestAPI()
}
