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
	"strconv"
	"strings"
	"time"
)

type FileHash struct {
    Path string `json:"path"`
    Hash     string `json:"hash"`
}

type Commit struct {
    ID int `json:"id"`
    Description string   `json:"description"`
    FileHash      []FileHash `json:"file_hash"`
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

func commits(description string, fileHashes []FileHash) error {
    commitsDir := filepath.Join(".gshot", "commits")
    if err := os.MkdirAll(commitsDir, 0755); err != nil {
        return err
    }

    commitFile := filepath.Join(commitsDir, "commits.json")

    var commits []Commit
    if data, err := os.ReadFile(commitFile); err == nil && len(data) > 0 {
        _ = json.Unmarshal(data, &commits)
    }

    var filteredFileHashes []FileHash
    
    if len(commits) == 0 {
        filteredFileHashes = fileHashes
    } else { 
        committed := make(map[string]struct{})
        for _, commit := range commits {
            for _, cf := range commit.FileHash { 
                committed[cf.Hash] = struct{}{}
            }
        }
 
        for _, fh := range fileHashes {
            if _, exists := committed[fh.Hash]; !exists {
                filteredFileHashes = append(filteredFileHashes, fh)
            }
        }
    }

    // no files changed
    if len(filteredFileHashes) <= 0 {
        fmt.Println("~ No files changed!")
        return nil
    }

    newID := 1
    if len(commits) > 0 {
        lastID := commits[len(commits) - 1].ID
        newID = lastID + 1
    } 

    newCommit := Commit{
        ID : newID,
        Description: description,
        FileHash:      filteredFileHashes,
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

func backToCommit(commitID string) error {
    commitsDir := filepath.Join(".gshot", "commits")
    if err := os.MkdirAll(commitsDir, 0755); err != nil {
        return err
    }

    commitFile := filepath.Join(commitsDir, "commits.json")

    var commits []Commit
    if data, err := os.ReadFile(commitFile); err == nil && len(data) > 0 {
        _ = json.Unmarshal(data, &commits)
    }

    var targetCommit *Commit 

    // find commit (use index to get pointer to slice element)
    for i := range commits {
        if strconv.Itoa(commits[i].ID) == commitID {
            targetCommit = &commits[i]
            break
        }
    }

    if targetCommit == nil {
        return fmt.Errorf("commit %s not found", commitID)
    }

    // ensure .gshot folder exists
    if err := os.MkdirAll(".gshot", 0755); err != nil {
        return err
    }

    // get files and overwrite
    for _, fileHash := range targetCommit.FileHash {
        if err := os.MkdirAll(filepath.Dir(fileHash.Path), 0755); err != nil {
            log.Println("Failed to create dir:", err)
            continue
        }

        srcPath := filepath.Join(".gshot/blobs/", fileHash.Hash)
        if err := OverwriteOrCreate(srcPath, fileHash.Path); err != nil {
            log.Println("Failed to overwrite file:", err)
        }
    }

    return nil
}


func OverwriteOrCreate(srcPath , dstPath string) error {
    src, err := os.Open(srcPath)

    if err != nil {
        return err
    }

    defer src.Close()

    dst , err := os.Create(dstPath)
    
    if err != nil {
        return err
    }

    defer dst.Close()

    _, err = io.Copy(dst,src)

    if err != nil {
        return err
    }

    return dst.Sync()
}

func printCommitsLog() {
    commitsDir := filepath.Join(".gshot", "commits")
    commitFile := filepath.Join(commitsDir, "commits.json")

    var commits []Commit
    if data, err := os.ReadFile(commitFile); err == nil && len(data) > 0 {
        if err := json.Unmarshal(data, &commits); err != nil {
            fmt.Println("❌ Failed to parse commits:", err)
            return
        }
    }

    if len(commits) == 0 {
        fmt.Println("📭 No commits found.")
        return
    }

    fmt.Println("📝 === Commit Log ===")
    for i, commit := range commits {
        fmt.Printf("\n✨ Commit #%d\n", i+1)
        fmt.Println(strings.Repeat("─", 30))
        fmt.Println("🗒️  Description:", commit.Description)
        fmt.Println("⏰ Timestamp:  ", commit.Timestamp)
        fmt.Println("🔗 Hashes:")
        for _, hash := range commit.FileHash {
            fmt.Println("  🟢", hash)
        }
    }
}

func main() {
    projectDir := "."

    ignoreDirs := []string{".gshot", "node_modules", ".git"}
    ignoreFiles := []string{"ignore.txt", "main.go", "go.mod"}

    files, err := getAllFiles(projectDir, ignoreDirs, ignoreFiles)
    if err != nil {
        log.Fatal(err)
    } 
    
    var filehash []FileHash
    for _, f := range files {
        fullPath := filepath.Join(projectDir, f)

        hash, err := storeBlob(fullPath)
        if err != nil {
            continue
        }

        fh := FileHash{
            Path: fullPath,
            Hash:     hash,
        }

        filehash = append(filehash, fh)
    }   
  
    commitMessage := flag.String("message", "", "commit message")
    showLog := flag.Bool("log", false, "show log message")  
    backTo := flag.String("back-to", "", "back to a commit by id")  

    flag.Parse()
 
    if *showLog {
        printCommitsLog()
        return
    }
 
    if *commitMessage != "" {
        if err := commits(*commitMessage, filehash); err != nil {
            fmt.Println("Error creating commit:", err)
        }
        return
    }  
 
    if *backTo != "" {
        backToCommit(*backTo)
        return
    }

    if err := initRepository(); err != nil {
        fmt.Println("Error initializing repository:", err)
        os.Exit(1)
    }
}