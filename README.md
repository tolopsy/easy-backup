# Easy Backup
Easy Backup is a CLI tool to backup files in your local computer into a dedicated directory.
Easy Backup is based on two main programs.

 1. `backup` program: It is responsible for management of records in the database. You can use the `backup` executable to add or remove files. You can also use the executable to list all the files to backup. In a nutshell, `backup` is the management tool.
 2. `backupd` program: This is the daemon responsible for the backup process.

You can find `backup` and `backupd` inside the [cmd folder](https://github.com/tolopsy/easy-backup/tree/main/cmd).
 
 ## How to setup Easy Backup
 All you need to setup Easy Backup is to run the make utility in your terminal as shown below
 ``` make ```
 This will build the daemon tool as well as the management tool for Easy Backup.
 Ensure that 
 - You are in the repo directory.
 - You have the make utility installed.


## Using Easy Backup
 
 ### How to run the management tool - `backup`
 
1. - To add path(s): 
`./dist/backup add /path/to/file1 ./path/to/file2`
Paths can be relative or absolute. And you can add as many paths as you want by separating them with a space.
	- To remove path(s):
	`./dist/backup remove ./path/to/file1 ./path/to/file2`
	- To list paths in db:
	`./dist/backup list`

	Note: The default location for the file-based db is ./data. But you can specify a custom location by using the -db flag when running the backup program.

	Just like this 👉🏾
	`./dist/backup -db=/custom/path/to/db [commands]`


### How to run the backup daemon - `backupd`

1. Simply run `./dist/backupd` to start the backup process.
    The backup daemon backs up into AWS S3 by default.
    The following environment variables are mandatory: S3Bucket, AWSRegion, AWSAccessKeyId and AWSSecretAccessKey
	- To use local storage instead of aws s3, set cloud flag to false. 👉🏾 `./dist/backupd -cloud=false`. The aforelisted environment variables are not needed for local storage.

	- **To specify a custom location for your file-based db**, append a db flag to the command.
	
		Like this 👉🏾
	`./dist/backupd -db=/path/to/customdb`
	
	- By default, the backup zip files will be stored in an folder named "backups" inside the s3 bucket (or project directory when using local storage). To specify a custom archive directory, append an archive flag to the command.
	
		Like this 👉🏾
	`./dist/backupd -archive=/path/to/archive`
	
	- The backup cycle runs every 10 seconds by default. This means that every 10 seconds, the daemon will perform the backup process. 

		You can update this cycle interval by appending an interval flag to the command.
		
		Like this 👉🏾
		`./dist/backupd -interval=40s`
	's' in '40s' above represents seconds. You can specify any interval you wish. 
	Use 'h' for hour and 'm' for minute. You can also combine the interval units like this - `./dist/backupd -interval=1h20m30s`.

    - To run backup once, include the `--once` flag.

Adios!
