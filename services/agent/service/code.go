package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"codev42-agent/model"
	"codev42-agent/storage"
	"codev42-agent/storage/repo"
	"codev42-agent/util"
)

type SaveCodeResult struct {
	Chunk           string
	FuncDeclaration string
	IsNew           bool
	IsUpdated       bool
}

func SaveCode(code string, filePath string, db *storage.RDBConnection) (map[int64]SaveCodeResult, error) {
	extension := filepath.Ext(filePath)
	keywords, err := util.GetKeywordsByExtension(extension)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return nil, err
	}

	// 키워드로 코드 분리
	chunks := util.SplitByKeywords(code, keywords)

	codeRepo := repo.NewCodeRepo(db)
	fileRepo := repo.NewFileRepo(db)
	var id int64
	file, err := fileRepo.GetFileByPath(context.Background(), filePath)
	if err != nil {
		if err.Error() == "record not found" {
			fileModel := &model.File{
				FilePath:  filePath,
				Directory: filepath.Dir(filePath),
			}
			id, err = fileRepo.InsertFile(context.Background(), fileModel)
			if err != nil {
				return nil, fmt.Errorf("failed to insert file: %v", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get file: %v", err)
		}
	} else {
		id = file.ID
	}

	var ret = make(map[int64]SaveCodeResult)
	for _, chunk := range chunks {
		if strings.TrimSpace(chunk) != "" {
			funcDeclaration := util.ExtractDeclaration(chunk, extension)
			chunkHash := util.HashChunk(chunk)
			code, err := codeRepo.GetCodeByFileIdAndName(context.Background(), id, funcDeclaration)
			newCodeModel := &model.Code{
				FileID:          id,
				FuncDeclaration: funcDeclaration,
				CodeChunk:       chunk,
				ChunkHash:       chunkHash,
			}
			if err != nil {
				if err.Error() == "record not found" {
					id, err := codeRepo.InsertCode(context.Background(), newCodeModel)
					if err != nil {
						return nil, fmt.Errorf("failed to insert code: %v", err)
					}
					ret[id] = SaveCodeResult{
						Chunk:           chunk,
						FuncDeclaration: funcDeclaration,
						IsNew:           true,
						IsUpdated:       false,
					}
				} else {
					return nil, fmt.Errorf("failed to get code: %v", err)
				}
			} else {
				isUpdated := code.ChunkHash != chunkHash
				if isUpdated {
					code.FuncDeclaration = funcDeclaration
					code.CodeChunk = chunk
					code.ChunkHash = chunkHash
					err := codeRepo.UpdateCode(context.Background(), code)
					if err != nil {
						return nil, fmt.Errorf("failed to update code: %v", err)
					}
				}
				ret[code.ID] = SaveCodeResult{
					Chunk:           chunk,
					FuncDeclaration: funcDeclaration,
					IsNew:           false,
					IsUpdated:       isUpdated,
				}
			}
		}
	}
	return ret, nil
}
