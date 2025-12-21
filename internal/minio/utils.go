package minio

func GetBucketNameByRef(ref string) string {
	return "baas-" + ref
}

func GetBucketPolicyNameByRef(ref string) string {
	return GetBucketNameByRef(ref) + "-policy"
}
