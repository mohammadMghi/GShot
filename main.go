package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Commit struct {
    Description string   `json:"description"`
    Hashes      []string `json:"hashes"`
    Timestamp   string   `json:"timestamp"`
}

type CommitTree struct{
	CommitFile string `json:"commit_file"`
}


func initRepository() error {
    vcsDir := ".gshot"
 
    dirs := []string{
        "commits",
        "blobs",
        "branches",
    }

    if _, err := os.Stat(vcsDir); err == nil { 
        return nil
    } else if !os.IsNotExist(err) { 
        return err
    }
 
    if err := os.MkdirAll(vcsDir, 0755); err != nil {
        return err
    }
 
    for _, dir := range dirs {
        path := filepath.Join(vcsDir, dir)
        if err := os.MkdirAll(path, 0755); err != nil {
            return err
        }
    } 

    headPath := filepath.Join(vcsDir, "HEAD")
    if err := os.WriteFile(headPath, []byte("branches/master"), 0644); err != nil {
        return err
    }
 
    masterBranch := filepath.Join(vcsDir, "branches", "master")
    if err := os.WriteFile(masterBranch, []byte(""), 0644); err != nil {
        return err
    }

    fmt.Println("Initialized Gshot! repository")
    return nil
}

func getAllFiles(root string, ignoreDirs []string, ignoreFiles []string) ([]string, error) {
    var files []string

    err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        } 

        if info.IsDir() {
            for _, dir := range ignoreDirs {
                if info.Name() == dir {
                    return filepath.SkipDir
                }
            }
            return nil
        }
 
        for _, f := range ignoreFiles {
            if info.Name() == f {
                return nil
            }
        }
 
        files = append(files, path)
        return nil
    })

    return files, err
}

func hashFile(filePath string) (string, error) {
    f, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer f.Close()

    hasher := sha256.New()
    if _, err := io.Copy(hasher, f); err != nil {
        return "", err
    }

    return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func commits(description string, hashes []string) error {
    commitsDir := filepath.Join(".gshot", "commits")
    if err := os.MkdirAll(commitsDir, 0755); err != nil {
        return err
    }

    commitFile := filepath.Join(commitsDir, "commits.json")

    var commits []Commit
    if data, err := os.ReadFile(commitFile); err == nil && len(data) > 0 {
        _ = json.Unmarshal(data, &commits)
    }
 
    if len(commits) > 0 {
        latestCommit := commits[len(commits)-1]

        // create a map for quick lookup
        existingHashes := make(map[string]struct{})
        for _, h := range latestCommit.Hashes {
            existingHashes[h] = struct{}{}
        }
 
        filteredHashes := []string{}
        for _, h := range hashes {
            if _, exists := existingHashes[h]; !exists {
                filteredHashes = append(filteredHashes, h)
            }
        }
        hashes = filteredHashes
    }

    // no files changed
    if len(hashes) <= 0 {
        fmt.Println("~ No files changed!")
        return nil
    }

    newCommit := Commit{
        Description: description,
        Hashes:      hashes,
        Timestamp:   time.Now().Format(time.RFC3339),
    }

    commits = append(commits, newCommit)

    data, err := json.MarshalIndent(commits, "", "  ")
    if err != nil {
        return err
    }

    if err := os.WriteFile(commitFile, data, 0644); err != nil {
        return err
    }

    fmt.Printf("✅ Commit added to: %s\n", commitFile)
    return nil
}

func removeAt(s []string, i int) []string {
    if i < 0 || i >= len(s) { 
        return s
    }
    return append(s[:i], s[i+1:]...)
}

func storeBlob(filePath string) (string, error) {
    hash, err := hashFile(filePath)
    if err != nil {
        return "", err
    }

    blobsDir := filepath.Join(".gshot", "blobs")
    if err := os.MkdirAll(blobsDir, 0755); err != nil {
        return "", err
    }

    destPath := filepath.Join(blobsDir, hash)
 
    if _, err := os.Stat(destPath); err == nil {
        return hash, nil
    }
 
    srcFile, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer srcFile.Close()

    destFile, err := os.Create(destPath)
    if err != nil {
        return "", err
    }
    defer destFile.Close()

    if _, err := io.Copy(destFile, srcFile); err != nil {
        return "", err
    }

    return hash, nil
}

func main() {
   	projectDir := "."  
	
	ignoreDirs := []string{".gshot", "node_modules" , ".git"} 
    ignoreFiles := []string{"ignore.txt" , "main.go" , "go.mod"}

    files, err := getAllFiles(projectDir,ignoreDirs,ignoreFiles)
    if err != nil {
        log.Fatal(err)
    }

	var hashes []string
 
	for _, f := range files {
		hash,err := storeBlob(f)

		if err != nil {

		}  
		hashes = append(hashes, hash) 
    }

    commitMessage := flag.String("message" , "" , "commit message")
    
    flag.Parse()

    if *commitMessage != "" {
        commits("this is first commit" ,hashes)
    } else {
        fmt.Println("Please set a message for your commit with --message flag")
        return
    }

    if err := initRepository(); err != nil {
        fmt.Println("Error initializing repository:", err)
        os.Exit(1)
    }
}