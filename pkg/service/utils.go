package service

import (
	"github.com/hstreamdb/deployment-tool/pkg/spec"
	"github.com/hstreamdb/deployment-tool/pkg/utils"
	"regexp"
	"strings"
)

func getMetaStoreUrl(tp spec.MetaStoreType, url string) string {
	switch tp {
	case spec.ZK:
		return "zk://" + url
	case spec.RQLITE:
		urls := strings.ReplaceAll(url, "http://", "")
		finalUrl := strings.Split(urls, ",")[0]
		return "rq://" + finalUrl
	case spec.Unknown:
		return ""
	}
	return ""
}

func needSeedNodes(version utils.Version) bool {
	return utils.CompareVersion(version, utils.Version082) > 0 && utils.CompareVersion(version, utils.Version084) != 0
}

func parseImage(imageStr string) (string, utils.Version) {
	reg := regexp.MustCompile(".*[:v]?\\d{1,3}.\\d{1,3}.\\d{1,3}")
	if !reg.MatchString(imageStr) {
		return imageStr, utils.Version{IsLatest: true}
	}

	fragment := strings.Split(imageStr, ":")
	image, version := fragment[0], fragment[1]
	return image, utils.CreateVersion(version)
}
