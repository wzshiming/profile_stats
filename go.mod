module github.com/wzshiming/profile_stats

go 1.16

require (
	github.com/ajstarks/svgo v0.0.0-20210406150507-75cfd577ce75
	github.com/google/go-github/v33 v33.0.0
	github.com/shurcooL/githubv4 v0.0.0-20201206200315-234843c633fa
	github.com/shurcooL/graphql v0.0.0-20200928012149-18c5c3165e3a // indirect
	github.com/vdobler/chart v0.0.0
	github.com/wzshiming/httpcache v0.1.0
	github.com/wzshiming/putingh v0.5.0
	github.com/wzshiming/xmlinjector v0.2.2
	golang.org/x/oauth2 v0.0.0-20210220000619-9bb904979d93
	golang.org/x/text v0.3.3
)

replace github.com/vdobler/chart => github.com/wzshiming/chart v1.0.0-fix
