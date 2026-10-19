package computeenvs

import (
	"encoding/json"
	"reflect"
	"testing"
)

func decode(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestToV2WriteModel_AwsBatch(t *testing.T) {
	v1 := `{"computeEnv":{"name":"ce","description":"d","credentialsId":"c1","platform":"aws-batch","workspaceId":1,"orgId":2,
		"config":{"discriminator":"aws-batch","region":"eu-west-1","workDir":"s3://b","nvnmeStorageEnabled":true,"lustreId":"l","storageType":"x",
		"forge":{"type":"SPOT","maxCpus":10}}},"labelIds":[5,6]}`
	got := decode(t, must(toV2WriteModel([]byte(v1))))
	want := map[string]any{
		"name": "ce", "description": "d", "credentialsId": "c1",
		"resourceLabelIds": []any{"5", "6"},
		"config": map[string]any{
			"platform": "AWS_BATCH", "region": "eu-west-1", "workDir": "s3://b", "nvmeStorageEnabled": true,
			"forge": map[string]any{"type": "SPOT", "maxCpus": float64(10)},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v\nwant %v", got, want)
	}
}

func TestToV2WriteModel_PlatformFromTopLevel(t *testing.T) {
	v1 := `{"computeEnv":{"name":"ce","platform":"azure-cloud","config":{"bootDiskSizeGB":50,"resourceGroup":"rg",
		"schedConfig":{"provisioningModel":"spotFirst","diskAllocation":"x"}}}}`
	got := decode(t, must(toV2WriteModel([]byte(v1))))
	cfg := got["config"].(map[string]any)
	if cfg["platform"] != "AZURE_CLOUD" || cfg["bootDiskSizeGb"] != float64(50) || cfg["resourceGroup"] != nil {
		t.Fatalf("config %v", cfg)
	}
	if sc := cfg["schedConfig"].(map[string]any); sc["provisioningModel"] != "SPOT_FIRST" || sc["diskAllocation"] != nil {
		t.Fatalf("schedConfig %v", sc)
	}
}

func TestToV2WriteModel_UnknownPlatform(t *testing.T) {
	if _, err := toV2WriteModel([]byte(`{"computeEnv":{"name":"ce","config":{"discriminator":"nope"}}}`)); err == nil {
		t.Fatal("expected an error")
	}
}

func TestToV1Describe(t *testing.T) {
	v2 := `{"id":"abc","name":"ce","description":null,"status":"AVAILABLE","primary":true,"credentialsId":"c1",
		"createdAt":"2026-10-19T10:00:00Z","updatedAt":"2026-10-19T11:00:00Z","lastUsedAt":null,
		"resourceLabels":[{"id":"7","name":"k","value":"v"}],
		"config":{"platform":"K8S","workDir":"/w","podCleanup":"ON_SUCCESS"}}`
	got := decode(t, must(toV1Describe([]byte(v2), 42)))["computeEnv"].(map[string]any)
	want := map[string]any{
		"id": "abc", "name": "ce", "status": "AVAILABLE", "primary": true, "credentialsId": "c1",
		"dateCreated": "2026-10-19T10:00:00Z", "lastUpdated": "2026-10-19T11:00:00Z",
		"workspaceId": float64(42), "platform": "k8s-platform",
		"labels": []any{map[string]any{"id": float64(7), "name": "k", "value": "v", "resource": true}},
		"config": map[string]any{"discriminator": "k8s-platform", "workDir": "/w", "podCleanup": "on_success"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v\nwant %v", got, want)
	}
}

func TestToV1Describe_AwsBatchAndSeqeraCompute(t *testing.T) {
	got := decode(t, must(toV1Describe([]byte(`{"id":"a","name":"n","config":{"platform":"AWS_BATCH","nvmeStorageEnabled":true}}`), 1)))["computeEnv"].(map[string]any)
	if cfg := got["config"].(map[string]any); cfg["nvnmeStorageEnabled"] != true || cfg["nvmeStorageEnabled"] != nil {
		t.Fatalf("config %v", cfg)
	}
	got = decode(t, must(toV1Describe([]byte(`{"id":"a","name":"n","config":{"platform":"SEQERA_COMPUTE","awsAccountId":"123"}}`), 1)))["computeEnv"].(map[string]any)
	if got["awsAccountId"] != "123" || got["config"].(map[string]any)["awsAccountId"] != nil {
		t.Fatalf("seqera compute %v", got)
	}
}

func TestToV2Patch(t *testing.T) {
	got := decode(t, must(toV2Patch([]byte(`{"name":"n","description":"","credentialsId":null}`))))
	want := map[string]any{"name": "n", "description": ""}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func must(b []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return b
}
