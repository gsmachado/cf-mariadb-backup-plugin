# cf-mariadb-backup-plugin

A Cloud Foundry CLI plugin to manage backups on service instances. Currently, only supporting MariaDB services.

Originally developed in the scope of [Swisscom AppCloud](http://developer.swisscom.com):
* [Swisscom Guid for Backups](https://docs.developer.swisscom.com/devguide-sc/services/backups.html)
* [Swisscom API for Service Instances](https://api.lyra-836.appcloud.swisscom.com/api-doc/#/Service_Instances)

## Installation

Currently you can only install this plugin through the binary:

1. Download the binary for your platform (Windows, Mac, or Linux) from the latest [Release](https://github.com/gsmachado/cf-mariadb-backup-plugin/releases)
2. Go to the directory where you downloaded the binary:
	* `cd path/to/downloaded/binary`
3. If you've already installed the plugin and is just updating, you must first run: **cf uninstall-plugin cf-mariadb-backup-plugin**
4. Then, install the plugin:
	* Windows: `cf install-plugin cf-mariadb-backup-plugin.exe`
	* Mac: `cf install-plugin cf-mariadb-backup-plugin_darwin`
	* Linux: `cf install-plugin cf-mariadb-backup-plugin_linux`
	* IMPORTANT: If you get a permission error on Mac or Linux, before installing it, run:
		* For Mac: `chmod +x cf-mariadb-backup-plugin_darwin`
		* For Linux: `chmod +x cf-mariadb-backup-plugin_linux`
5. Verify the plugin installed by looking for it:
	* `cf plugins`

## Usage

This plugin currently supports the following commands.

#### Listing the backups

```
cf list-mariadb-backups -s <SERVICE_NAME>
```

where the `<SERVICE_NAME>` should be a MariaDB service instance.

#### Create a backup

```
cf create-mariadb-backup -s <SERVICE_NAME>
```

Also, there's the possibility to specify the max amount of backups allowed before automatically rotating them:

```
cf create-mariadb-backup -s <SERVICE_NAME> -m <MAX_BACKUPS_ROTATION>
```

More specifically, this command will just create a backup if the total amount of backups on such service is below the `<MAX_BACKUPS_ROTATION>` value. If the amount exceeds the `<MAX_BACKUPS_ROTATION>`, then the *oldest backup will be deleted*, automatically, before creating a new backup. This is very useful if your Cloud Foundry provider limits the amount of backups per service.

#### Delete a backup

```
cf delete-mariadb-backup -s <SERVICE_NAME> -b <BACKUP_GUID>
```

The `<BACKUP_GUID>` can be obtained running the listing backup command.

## Development

To be described.

## ToDo's

* Restore command to restore a specific backup
* Implement Go tests
* Describe how to build and some implementation details on the README.md