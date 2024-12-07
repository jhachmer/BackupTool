# BackupTool

This is a small backup tool configurable with a YAML file.

## Config
Just create a YAMl file called `dirs.yaml` with the following structure:
- `destination` is where your zip-files will be stored
- under `jobs` put in your desired backups
  - the `name` field is the name of the created ZIP-file.
  - with the `dirs` list you can specify one more folder locations 
```yaml
# all dirs need absolute paths
# e.g. E:\Backups
destination: E:\Backups #Your Backup Location
jobs:
  - name: WoW
    dirs:
      - D:\World of Warcraft\_retail_\WTF
      - D:\World of Warcraft\_retail_\Interface
  - name: Videos
    dirs:
      - E:\Videos\
  - name: Other
    dirs:
      - /Some/other/dir
```

## Run
Compile with `go build main.go` and execute the resulting binary after configuring