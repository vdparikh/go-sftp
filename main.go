package main

import (
	"os"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jasonlvhit/gocron"
	sftp "github.com/pkg/sftp"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

var (
	sftpClient *sftp.Client
)

func init() {
}

func downloadRemoteFile(remotePath string, localPath string) {
	log.Infof("Listing Directory %s", remotePath)
	w, _ := sftpClient.ReadDir(remotePath)

	if len(w) == 0 {
		log.Warningf("No Fies Found in Directory %s", remotePath)
		return
	}
	for _, file := range w {
		log.Debug(file.Name())
		if !file.IsDir() {
			startTime := time.Now()

			remoteFilePath := remotePath + "/" + file.Name()
			log.Infof("Requesting File %s From Remote Server", remoteFilePath)

			srcFile, err := sftpClient.Open(remoteFilePath)
			if err != nil {
				thorwError("Unable to Open Remote File", err)
				return
			}
			defer srcFile.Close()

			inputFilePrefix := time.Now().Format("2006_01_02_15_04_05") + "_"
			localFileName := localPath + "/" + inputFilePrefix + path.Base(remoteFilePath)

			dstFile, err := os.Create(path.Join(localFileName))
			if err != nil {
				thorwError("Unable to Create Local File", err)
				return
			}
			defer dstFile.Close()

			if _, err = srcFile.WriteTo(dstFile); err != nil {
				thorwError("Unable to Write to Local File", err)
				return
			}

			log.Infof("File %s Downloaded locally as %s from remote server in %v", remoteFilePath, localFileName, time.Since(startTime))
		}
	}
}

func downloadFiles() {
	// Dial
	log.Infof("Cron Schedule started at %v", time.Now())
	log.Infof("Connecting to %s", viper.GetString("ftp.server"))

	sshConfig := &ssh.ClientConfig{
		User:            viper.GetString("ftp.user"),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(viper.GetString("ftp.password")),
		},
	}

	sshConn, err := ssh.Dial("tcp", viper.GetString("ftp.server"), sshConfig)
	if err != nil {
		thorwError("Unable to Connect", err)
		return
	}
	defer sshConn.Close()
	log.Infof("Connected to %s", viper.GetString("ftp.server"))

	sftpClient, err = sftp.NewClient(sshConn)
	defer sftpClient.Close()

	downloadRemoteFile(viper.GetString("ftp.location.remotePath"), viper.GetString("ftp.location.localPath"))

	log.Infof("Cron Schedule completed at %v", time.Now())
}

func main() {

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()

	if err != nil {
		panic("Failed Reading Config File! ")
	}

	// Log as Text
	log.SetOutput(os.Stdout)
	if viper.GetString("log.output") == "file" {
		f, err := os.OpenFile("log.out", os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			log.Fatal("Unable to open Log file for writing")
		}
		log.SetOutput(f)
	}

	log.SetFormatter(&log.TextFormatter{})

	lvl, err := log.ParseLevel(viper.GetString("log.logLevel"))
	if err != nil {
		log.Fatal("Invalid Logging Level")
	}
	log.SetLevel(lvl)

	log.Infof("Starting Cron Schedule to run every %d Minutes", uint64(viper.GetInt64("cron.minutes")))
	s := gocron.NewScheduler()
	s.Every(uint64(viper.GetInt64("cron.minutes"))).Minutes().Do(downloadFiles)
	<-s.Start()
}

func thorwError(msg string, err error) {
	// Send email
	// Leverage something else
	log.WithFields(log.Fields{
		"error": err.Error(),
	}).Fatal(msg)
}
