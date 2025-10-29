resource "seqera_bitbucket_credential" "my_bitbucketcredential" {
  base_url     = "https://bitbucket.org/myorg"
  name         = "...my_name..."
  token        = "ATBB..."
  username     = "myuser@example.com"
  workspace_id = 10
}