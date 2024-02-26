//go:generate oapi-codegen -package v1 -generate types,chi-server,strict-server,spec,skip-prune -o v1.gen.go ../../mediaserver-v1.yml
package v1
