# Easy Backup
Easy Backup is a CLI tool to backup files in your local computer into a dedicated directory.
Easy Backup is based on two main programs.

 1. `backup` program: it contains a NoSQL-like file-based database that stores the location of files/directories you want to backup. It is also responsible for management of records in the database. You can use the `backup` executable to add or remove paths from the Path collection. You can also use the executable to list all the paths (i.e location of files/folders to backup). In a nutshell, `backup` is the management tool.
 2. `backupd` program: This is the daemon responsible for the backup process. By default, the dedicated archive directory (called backups) is in the `backupd` folder.

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
 

 1. Change directory to the [management directory](https://github.com/tolopsy/easy-backup/tree/main/cmd/backup).
	 `cd ./cmd/backup`
2. - To add path(s): 
`./backup add /path/to/file ./path/to/folder`
Paths can be relative or absolute. And you can add as many paths as you want by separating them with a space.
	- To remove path(s):
	`./backup remove ./path/to/file1 /path/to/folder2`
	- To list paths in db:
	`./backup list`

	Note: The default location for the file-based db is ./cmd/backup/data. But you can specify a custom location by using the -db flag when running the backup program.

	Just like this 👉🏾
	`./backup -db=/custom/path/to/db [commands]`
	The paths are stored inside `Paths.filedb`.


### How to run the backup daemon - `backupd`
1. Change directory to the [daemon directory](https://github.com/tolopsy/easy-backup/tree/main/cmd/backupd)
`cd ./cmd/backupd`
2. Simply run `./backupd` to start the backup process.
	
	- **To specify a custom location for your file-based db**, append a db flag to the command.
	
		Like this 👉🏾
	`./backupd -db=/path/to/customdb`
	
	- By default, the backup zip files will be stored in an archive directory named "backups" inside the daemon directory. To specify a custom archive directory, append an archive flag to the command.
	
		Like this 👉🏾
	`./backupd -archive=/path/to/archive`
	
	- The backup cycle runs every 10 seconds by default. This means that every 10 seconds, the daemon will perform the backup process. 

		You can update this cycle interval by appending an interval flag to the command.
		
		Like this 👉🏾
		`./backupd -interval=40s`
	's' in '40s' above represents seconds. You can specify any interval you wish. 
	Use 'h' for hour and 'm' for minute. You can also combine the interval units like this - `./backupd -interval=1h20m30s`.

    - To run backup once, include the `--once` flag.

Adios!
