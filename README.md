# GShot
GShot has an educational aspect.
You can read it and get inspired by how a version control works.
You can clone it out of the box and run the following instruction:

For commit something
```
  go main.go --message "commit message"
```
It creates a hash behind the since and and allocated to a file and put generated file inside ```.ghost/blobs``` files then creating a commits.json and stores your commit with message inside it as a json.

If you need to see logs (commits):
```
  go main.go --logs
```

If you want to go backward and forward between your commits
```
  go main.go --back-to id
```

I'll add more features and write a clear documentation on how it works.
