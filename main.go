package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"errors"
	"strconv"
	"github.com/gsmachado/cf-mariadb-backup-plugin/models"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/plugin"
)

type MariaDBBackupPlugin struct{}

const (
	AuthorEmail string = "gsm@machados.org"
	GitHubProjectURL string = "http://github.com/gsmachado/cf-mariadb-backup-plugin"
)

func (c *MariaDBBackupPlugin) Run(cliConnection plugin.CliConnection, args []string) {	
	switch args[0] {
		case "list-mariadb-backups":
			commandListBackups(args, cliConnection)
		case "create-mariadb-backup":
			commandCreateBackup(args, cliConnection)
		case "delete-mariadb-backup":
			commandDeleteBackup(args, cliConnection)
		case "CLI-MESSAGE-UNINSTALL":
			printGoodbyeBanner()
	}
}

func getCurrentOrgString(cliConnection plugin.CliConnection) (string) {
	org, err := cliConnection.GetCurrentOrg()
	if err != nil {
		exit1("Error getting the current organization.")
	}
	return org.Name
}

func getCurrentSpaceString(cliConnection plugin.CliConnection) (string) {
	space, err := cliConnection.GetCurrentSpace()
	if err != nil {
		exit1("Error getting the current space.")
	}
	return space.Name
}

func getBackups(cliConnection plugin.CliConnection, serviceInstance string) (model.ServiceInstanceResults, error) {
	nextURL := "/custom/service_instances/" + serviceInstance + "/backups"
	serviceInstanceResults := model.ServiceInstanceResults{}

	for nextURL != "" {
		output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", nextURL)
		if err != nil {
			return model.ServiceInstanceResults{}, err
		}
		// joining since it's an array of strings
		outputStr := strings.Join(output, "")
		outputBytes := []byte(outputStr)

		results := model.ServiceInstanceResults{}
		err = json.Unmarshal(outputBytes, &results)
		if err != nil {
			return model.ServiceInstanceResults{}, err
		}

		serviceInstanceResults.Resources = append(serviceInstanceResults.Resources, results.Resources...)
		serviceInstanceResults.TotalResults = results.TotalResults;
		serviceInstanceResults.TotalPages = results.TotalPages;

		if results.NextURL != nil {
			nextURL = *results.NextURL
		} else {
			nextURL = ""
		}
	}
	return serviceInstanceResults, nil
}

func createBackup(cliConnection plugin.CliConnection, serviceInstance string) (bool, error) {
	result := model.ServiceInstanceBackup{}
	url := "/custom/service_instances/" + serviceInstance + "/backups"
	output, err1 := cliConnection.CliCommandWithoutTerminalOutput("curl", "-X", "POST", url)
	if err1 != nil {
		return false, err1
	}
	// joining since it's an array of strings
	outputStr := strings.Join(output, "")
	outputBytes := []byte(outputStr)
	err2 := json.Unmarshal(outputBytes, &result)
	if err2 != nil || result.Entity == nil {
		return false, errors.New("Create backup command was not successful.\nDetails:\n" + output[0])
	}
	return true, nil
}

func getBackup(cliConnection plugin.CliConnection, serviceInstance string, backupInstance string) (model.ServiceInstanceBackup, error) {
	result := model.ServiceInstanceBackup{}
	url := "/custom/service_instances/" + serviceInstance + "/backups/" + backupInstance
	output, err1 := cliConnection.CliCommandWithoutTerminalOutput("curl", url)
	if err1 != nil {
		return result, err1
	}
	// joining since it's an array of strings
	outputStr := strings.Join(output, "")
	outputBytes := []byte(outputStr)
	err2 := json.Unmarshal(outputBytes, &result)
	if err2 != nil {
		return model.ServiceInstanceBackup{}, err2
	}
	return result, nil
}

func getService(cliConnection plugin.CliConnection, serviceName string) (plugin_models.GetService_Model) {
	service, err := cliConnection.GetService(serviceName)
	if err != nil {
		exit1("Error getting service: " + err.Error())
	}
	if len(service.Guid) == 0 {
		exit1("Does the service exists?")
	}
	if !strings.Contains(service.ServiceOffering.Name, "mariadb") {
		exit1("The service " + serviceColor(service.Name) + " is not a MariaDB instance. This plugin does not yet support the backup of other service instances.")
	}
	return service
}

func deleteBackup(cliConnection plugin.CliConnection, serviceInstance string, backupInstance string) (bool, error) {
	url := "/custom/service_instances/" + serviceInstance + "/backups/" + backupInstance
	_, err := cliConnection.CliCommandWithoutTerminalOutput("curl", "-X", "DELETE", url)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *MariaDBBackupPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "cf-mariadb-backup-plugin",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 1,
			Build: 0,
		},
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 7,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "list-mariadb-backups",
				HelpText: "List all backups of a specific MariaDB service.",
				UsageDetails: plugin.Usage{
					Usage: "list-mariadb-backups:\n   cf list-mariadb-backups -s SERVICE_NAME",
				},
			},
			{
				Name:     "create-mariadb-backup",
				HelpText: "Create a backup of a specific MariaDB service. You can specify the max amount of backups before rotation (delete the oldest).",
				UsageDetails: plugin.Usage{
					Usage: "create-mariadb-backup:\n   cf create-mariadb-backup -s SERVICE_NAME [-m MAX_BACKUPS_ROTATION]",
				},
			},
			{
				Name:     "delete-mariadb-backup",
				HelpText: "Delete the backup of a specific MariaDB service.",
				UsageDetails: plugin.Usage{
					Usage: "delete-mariadb-backup:\n   cf delete-mariadb-backup -s SERVICE_NAME -b BACKUP_GUID",
				},
			},
		},
	}
}

func parseArgumentsServiceName(args []string) (flags.FlagContext, error) {
	fc := flags.New()
	fc.NewStringFlag("service-name", "s", "Service name, e.g., --service-name SERVICE_NAME")
	err := fc.Parse(args...)
	if !fc.IsSet("service-name") {
		exit1("Error parsing the argument " + redFgColor("--service-name (-s)") + ". Did you forget to specify it?")
	}
	if err != nil {
		exit1("Error parsing the arguments: " + err.Error())
	}
	return fc, err
}

func parseArgumentsServiceNameAndRotationOption(args []string) (flags.FlagContext, error) {
	fc := flags.New()
	fc.NewStringFlag("service-name", "s", "Service name, e.g., --service-name SERVICE_NAME")
	fc.NewIntFlag("max-backups-rotation", "m", "Delete the oldest backup when the max amount of backups is reached, e.g., --max-backups-rotation AMOUNT")
	err := fc.Parse(args...)
	if !fc.IsSet("service-name") {
		exit1("Error parsing the argument " + redFgColor("--service-name (-s)") + ". Did you forget to specify it?")
	}
	if err != nil {
		exit1("Error parsing the arguments: " + err.Error())
	}
	return fc, err
}

func parseArgumentsServiceNameAndBackupGuid(args []string) (flags.FlagContext, error) {
	fc := flags.New()
	fc.NewStringFlag("service-name", "s", "Service name, e.g., --service-name SERVICE_NAME")
	fc.NewStringFlag("backup-guid", "b", "Backup GUID, e.g., --backup-guid BACKUP_GUID")
	err := fc.Parse(args...)
	if !fc.IsSet("service-name") || !fc.IsSet("backup-guid") {
		exit1("Error parsing the arguments " + redFgColor("--service-name (-s)") + " and " + redFgColor("--backup-guid (-b)") + ". Did you forget to specify it?")
	}
	if err != nil {
		exit1("Error parsing the arguments: " + err.Error())
	}
	return fc, err
}

func exit1(err string) {
	fmt.Println(redFgColor("FAILED") + "\n" + err + "\n")
	os.Exit(1)
}

func printServiceInstanceResultsAsJSON(backupResults model.ServiceInstanceResults) {
	out, err := json.MarshalIndent(backupResults, "", "   ")
    if err != nil {
       panic (err)
    }
    fmt.Println(string(out))
}

func printServiceInstanceBackupAsJSON(backup model.ServiceInstanceBackup) {
	out, err := json.MarshalIndent(backup, "", "   ")
    if err != nil {
       panic (err)
    }
    fmt.Println(string(out))
}

func getOldestBackup(backupResults model.ServiceInstanceResults) (model.ServiceInstanceBackup) {
	if len(backupResults.Resources) < 1 {
		return model.ServiceInstanceBackup{}
	}
	var oldest = model.ServiceInstanceBackup{}
	for _, e := range backupResults.Resources {
		if oldest.Metadata == nil {
			oldest = e
		} else {
			if e.Metadata.CreatedAt.Before(oldest.Metadata.CreatedAt) {
				oldest = e;
			}
		}
	}
	return oldest
}

func commandCreateBackup(args []string, cliConnection plugin.CliConnection) {
	fc, err := parseArgumentsServiceNameAndRotationOption(args)
	if err != nil {
		exit1(err.Error())
	}
	serviceName := fc.String("service-name")
	service := getService(cliConnection, serviceName)
	
	if fc.IsSet("max-backups-rotation") {
		backupResults, err := getBackups(cliConnection, service.Guid)
		if err != nil {
			exit1("Error getting backups for service '" + serviceColor(service.Name) + "': " + err.Error())
		}
		if len(backupResults.Resources) >= fc.Int("max-backups-rotation") {
			strAmountBackups := strconv.Itoa(len(backupResults.Resources))
			strMaxBackups := strconv.Itoa(fc.Int("max-backups-rotation"))
			fmt.Printf("Currently there are %s backups in service %s. The max specified amount of backups is %s.\n\n", serviceColor(strAmountBackups), serviceColor(service.Name), serviceColor(strMaxBackups))
			oldestBackup := getOldestBackup(backupResults)
			backupGuid := oldestBackup.Metadata.GUID
			fmt.Printf("Deleting the oldest backup in org %s / space %s of service %s...\n", serviceColor(getCurrentOrgString(cliConnection)), serviceColor(getCurrentSpaceString(cliConnection)), serviceColor(service.Name))
			_, err = deleteBackup(cliConnection, service.Guid, backupGuid)
			if err != nil {
				exit1("Error deleting the backup " + boldText(backupGuid) + " of service " + serviceColor(serviceName) + ": " + err.Error())
			}
			fmt.Printf("%s\n\n", hiGreenFgColor("OK"))
		}
	}
    fmt.Printf("Creating a backup in org %s / space %s of service %s...\n", serviceColor(getCurrentOrgString(cliConnection)), serviceColor(getCurrentSpaceString(cliConnection)), serviceColor(service.Name))
	_, err2 := createBackup(cliConnection, service.Guid)
	if err2 != nil {
		exit1("Error creating a backup for service '" + serviceColor(service.Name) + "': " + err2.Error())
	}
	fmt.Printf("%s\n\n", hiGreenFgColor("OK"))
}

func commandDeleteBackup(args []string, cliConnection plugin.CliConnection) {
	fc, err := parseArgumentsServiceNameAndBackupGuid(args)
	if err != nil {
		exit1(err.Error())
	}
	serviceName := fc.String("service-name")
	backupGuid := fc.String("backup-guid")
	service := getService(cliConnection, serviceName)
	_, errGetBackup := getBackup(cliConnection, service.Guid, backupGuid)
	if errGetBackup != nil {
		exit1("Error finding the backup " + boldText(backupGuid) + " of service " + serviceColor(serviceName) + ": " + err.Error())
	}
    fmt.Printf("Deleting a backup in org %s / space %s of service %s...\n", serviceColor(getCurrentOrgString(cliConnection)), serviceColor(getCurrentSpaceString(cliConnection)), serviceColor(service.Name))
	_, err = deleteBackup(cliConnection, service.Guid, backupGuid)
	if err != nil {
		exit1("Error deleting the backup " + boldText(backupGuid) + " of service " + serviceColor(serviceName) + ": " + err.Error())
	}
	fmt.Printf("%s\n\n", hiGreenFgColor("OK"))
}

func commandListBackups(args []string, cliConnection plugin.CliConnection) {
	fc, err := parseArgumentsServiceName(args)
	if err != nil {
		exit1(err.Error())
	}
	serviceName := fc.String("service-name")
	service := getService(cliConnection, serviceName)
    fmt.Printf("Listing backups in org %s / space %s of service %s...\n", serviceColor(getCurrentOrgString(cliConnection)), serviceColor(getCurrentSpaceString(cliConnection)), serviceColor(service.Name))
	backupResults, err := getBackups(cliConnection, service.Guid)
	if err != nil {
		exit1("Error getting backups for service '" + serviceColor(service.Name) + "': " + err.Error())
	}
	fmt.Printf("%s\n\n", hiGreenFgColor("OK"))
	printBackups(service.Name, backupResults)
	printRestores(service.Name, backupResults)
}


func printBackups(serviceName string, backupResults model.ServiceInstanceResults) {
	fmt.Printf("Backups of %s:\n", serviceColor(serviceName))	
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Index", "Backup GUID", "Backup Date Created", "Backup Status"})

	backups := backupResults.Resources;

	for i, e := range backups {
		backupGUID := e.Metadata.GUID
		backupDateCreated := e.Metadata.CreatedAt
		backupStatus := string(e.Entity.Status)
		switch backupStatus {
			case string(model.CreateSucceeded):
				backupStatus = successfulColor(backupStatus)
			case string(model.CreateInProgress):
				backupStatus = warningColor(backupStatus)
			default:
				backupStatus = failureColor(backupStatus)
		}
		tableElement := []string{strconv.Itoa(i+1), backupGUID, backupDateCreated.Format("2006-01-02 15:04:05"), backupStatus}
		table.Append(tableElement)
	}
	table.Render()
	fmt.Printf("\n")
}

func printRestores(serviceName string, backupResults model.ServiceInstanceResults) {
	fmt.Printf("Restores of %s:\n", serviceColor(serviceName))
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Index", "Backup GUID", "Restore GUID", "Restore Date Created", "Restore Status"})
	backups := backupResults.Resources;
	for i1, e1 := range backups {
		for _, e2 := range e1.Entity.Restores {
			backupGUID := e1.Metadata.GUID
			restoreGUID := e2.Metadata.GUID
			restoreDateCreated := e2.Metadata.CreatedAt
			restoreStatus := string(e2.Entity.Status)
			switch restoreStatus {
				case string(model.Succeeded):
					restoreStatus = successfulColor(restoreStatus)
				default:
					restoreStatus = failureColor(restoreStatus)
			}
			tableElement := []string{strconv.Itoa(i1+1), backupGUID, restoreGUID, restoreDateCreated.Format("2006-01-02 15:04:05"), restoreStatus}
			table.Append(tableElement)
		}
	}
	table.Render()
	fmt.Printf("\n")
}

func printGoodbyeBanner() {
	fmt.Println()
	fmt.Println(boldText("Thanks for using cf-mariadb-backup-plugin!"))
	fmt.Println("Send some feedback to: ")
	fmt.Println("- " + redFgColor(AuthorEmail))
	fmt.Println("- " + redFgColor(GitHubProjectURL))
	fmt.Println()
}

func serviceColor(param string) (string) {
	color := color.New(color.FgHiCyan).SprintFunc()
	return color(param)
}

func successfulColor(param string) (string) {
	color := color.New(color.BgGreen, color.FgBlack).SprintFunc()
	return color(param)
}

func warningColor(param string) (string) {
	color := color.New(color.BgYellow, color.FgBlack).SprintFunc()
	return color(param)
}

func failureColor(param string) (string) {
	color := color.New(color.BgRed, color.FgBlack).SprintFunc()
	return color(param)
}

func redFgColor(param string) (string) {
	color := color.New(color.FgRed).SprintFunc()
	return color(param)
}

func boldText(param string) (string) {
	color := color.New(color.Bold).SprintFunc()
	return color(param)
}

func hiGreenFgColor(param string) (string) {
	color := color.New(color.FgHiGreen).SprintFunc()
	return color(param)
}

func main() {
	plugin.Start(new(MariaDBBackupPlugin))
}