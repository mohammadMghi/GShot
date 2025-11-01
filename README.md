# GShot
GShot has an educational aspect.
You can read it and get inspired by how a version control works.
You can clone it out of the box and run the following instruction:

## Init
```
  go run main.go --init
```
This command creates a .gshot directory inside the root directory. It contains blobs, branches, and commits.
What are blobs?
They are the storage of files. After encrypting and calculating the SHA256, we store the file with its SHA256 name inside blobs.
Why do this?
Because we need to access the files by their SHA256. If one or more files change, we create new commits and compare them with the last SHA256. If a file stays the same and nothing changes inside it, the SHA256 remains the same as the previous one, so we don't need to store new files inside blobs or change any commit in .gshot/commits/commits.json. But if something changes inside our files, we have to calculate the SHA256 again and store a new commit inside commits.json.

### code documentation

``` initRepository() ``` this a function is represent for doing this.

## Commit
```
  go run main.go --commit "commit message"
```
### code documentation
Checks if branch exists or not , if not it going to create a master branch


It creates a hash behind the since and and allocated to a file and put generated file inside ```.ghost/blobs``` files then creating a commits.json and stores your commit with message inside it as a json.

## Log
If you need to see logs (commits):
```
  go run main.go --logs
```
## Change commit
If you want to go backward and forward between your commits
```
  go run main.go --back-to id
```

