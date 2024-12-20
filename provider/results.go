package provider

import (
	"github.com/couchbase/gocb/v2"

	// Generated bindings
	"github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings/exports/wasmcloud/couchbase/document"
)

// Result transformers for the document API.

func GetAllReplicasResult(result *gocb.GetAllReplicasResult) []*document.DocumentGetReplicaResult {
	var replicaResults []*document.DocumentGetReplicaResult
	next := result.Next()
	for next != nil {
		replicaResult := GetReplicaResult(next)
		replicaResults = append(replicaResults, &replicaResult)
		next = result.Next()
	}
	return replicaResults
}

func GetReplicaResult(result *gocb.GetReplicaResult) document.DocumentGetReplicaResult {
	return document.DocumentGetReplicaResult{
		Cas:       uint64(result.Result.Cas()),
		IsReplica: result.IsReplica(),
	}
}

func GetResult(result *gocb.GetResult) (document.DocumentGetResult, error) {
	var content string
	err := result.Content(&content)
	if err != nil {
		return document.DocumentGetResult{}, err
	}
	expiresInNs := uint64(result.ExpiryTime().Nanosecond())
	returnDoc := document.Document{}
	returnDoc.SetRaw(string(content))
	return document.DocumentGetResult{
		Document:    &returnDoc,
		ExpiresInNs: &expiresInNs,
		Cas:         uint64(result.Cas()),
	}, nil
}

func MutationMetadata(metadata *gocb.MutationResult) document.MutationMetadata {
	return document.MutationMetadata{
		Cas:           uint64(metadata.Cas()),
		Bucket:        metadata.MutationToken().BucketName(),
		PartitionId:   metadata.MutationToken().PartitionID(),
		PartitionUuid: metadata.MutationToken().PartitionUUID(),
		Seq:           metadata.MutationToken().SequenceNumber(),
	}
}
