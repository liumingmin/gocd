package gocd

type CdS3Info struct {
	s3AK         string
	s3SK         string
	s3Endpoint   string
	s3Bucket     string
	s3Region     string
	s3GetToolUrl string
}

func NewCdS3Info(s3AK, s3SK, s3Endpoint, s3Bucket, s3Region, s3getToolUrl string) *CdS3Info {
	return &CdS3Info{
		s3AK:         s3AK,
		s3SK:         s3SK,
		s3Endpoint:   s3Endpoint,
		s3Bucket:     s3Bucket,
		s3Region:     s3Region,
		s3GetToolUrl: s3getToolUrl,
	}
}

func (s *CdS3Info) envVar() map[string]string {
	return map[string]string{
		"GOCD_S3_AK":       s.s3AK,
		"GOCD_S3_SK":       s.s3SK,
		"GOCD_S3_ENDPOINT": s.s3Endpoint,
		"GOCD_S3_BUCKET":   s.s3Bucket,
		"GOCD_S3_REGION":   s.s3Region,
	}
}
