// Package semverx 基于 github.com/Masterminds/semver/v3 封装的常用语义化版本功能。
// 输入输出统一使用 string（内部转换为 semver.Version），便于业务代码直接使用。
// 版本号规则（SemVer 2.0.0）：主版本号.次版本号.修订号[-预发布版本][+构建元数据]
package semverx

import (
	"fmt"
	"sort"
	"strings"

	semver "github.com/Masterminds/semver/v3"
)

// Parse 宽松解析版本号：支持 "1"、"1.2"、"1.2.3-beta+build"，允许 "v" 前缀。
// 缺失的次版本/修订号按 0 补齐，如 "v1.2" 解析为 1.2.0。
func Parse(v string) (*semver.Version, error) {
	sv, err := semver.NewVersion(v)
	if err != nil {
		return nil, fmt.Errorf("版本号格式不合法：%w，输入：%s", err, v)
	}
	return sv, nil
}

// StrictParse 严格解析版本号：必须为主版本.次版本.修订号三段，可带 "v" 前缀。
func StrictParse(v string) (*semver.Version, error) {
	sv, err := semver.StrictNewVersion(v)
	if err != nil {
		return nil, fmt.Errorf("版本号格式不合法(严格)：%w，输入：%s", err, v)
	}
	return sv, nil
}

// IsValid 判断是否为合法版本号（宽松模式）。
func IsValid(v string) bool {
	_, err := semver.NewVersion(v)
	return err == nil
}

// Normalize 规范化版本号，返回完整三段格式，如 "v1.2" → "1.2.0"。
func Normalize(v string) (string, error) {
	sv, err := Parse(v)
	if err != nil {
		return "", err
	}
	return sv.String(), nil
}

// Compare 比较两个版本号：a>b 返回 1，a==b 返回 0，a<b 返回 -1。
// 比较遵循 SemVer 规则：预发布版本低于正式版本，构建元数据不参与比较。
func Compare(a, b string) (int, error) {
	va, vb, err := parsePair(a, b)
	if err != nil {
		return 0, err
	}
	return va.Compare(vb), nil
}

// Equal 判断两个版本号是否相等（忽略构建元数据）。
func Equal(a, b string) (bool, error) {
	va, vb, err := parsePair(a, b)
	if err != nil {
		return false, err
	}
	return va.Equal(vb), nil
}

// GreaterThan 判断 a 是否大于 b。
func GreaterThan(a, b string) (bool, error) {
	va, vb, err := parsePair(a, b)
	if err != nil {
		return false, err
	}
	return va.GreaterThan(vb), nil
}

// LessThan 判断 a 是否小于 b。
func LessThan(a, b string) (bool, error) {
	va, vb, err := parsePair(a, b)
	if err != nil {
		return false, err
	}
	return va.LessThan(vb), nil
}

// GreaterThanOrEqual 判断 a 是否大于等于 b。
func GreaterThanOrEqual(a, b string) (bool, error) {
	va, vb, err := parsePair(a, b)
	if err != nil {
		return false, err
	}
	return va.Compare(vb) >= 0, nil
}

// LessThanOrEqual 判断 a 是否小于等于 b。
func LessThanOrEqual(a, b string) (bool, error) {
	va, vb, err := parsePair(a, b)
	if err != nil {
		return false, err
	}
	return va.Compare(vb) <= 0, nil
}

// Major 获取主版本号。
func Major(v string) (uint64, error) {
	sv, err := Parse(v)
	if err != nil {
		return 0, err
	}
	return sv.Major(), nil
}

// Minor 获取次版本号。
func Minor(v string) (uint64, error) {
	sv, err := Parse(v)
	if err != nil {
		return 0, err
	}
	return sv.Minor(), nil
}

// Patch 获取修订号。
func Patch(v string) (uint64, error) {
	sv, err := Parse(v)
	if err != nil {
		return 0, err
	}
	return sv.Patch(), nil
}

// Prerelease 获取预发布版本信息，如 "1.2.3-beta.1" → "beta.1"，无则返回空串。
func Prerelease(v string) (string, error) {
	sv, err := Parse(v)
	if err != nil {
		return "", err
	}
	return sv.Prerelease(), nil
}

// Metadata 获取构建元数据，如 "1.2.3+20130313" → "20130313"，无则返回空串。
func Metadata(v string) (string, error) {
	sv, err := Parse(v)
	if err != nil {
		return "", err
	}
	return sv.Metadata(), nil
}

// IncMajor 主版本号加 1，并将次版本、修订号、预发布版本清零，如 "1.2.3" → "2.0.0"。
func IncMajor(v string) (string, error) {
	sv, err := Parse(v)
	if err != nil {
		return "", err
	}
	return sv.IncMajor().String(), nil
}

// IncMinor 次版本号加 1，并将修订号、预发布版本清零，如 "1.2.3" → "1.3.0"。
func IncMinor(v string) (string, error) {
	sv, err := Parse(v)
	if err != nil {
		return "", err
	}
	return sv.IncMinor().String(), nil
}

// IncPatch 修订号加 1；无预发布版本时如 "1.2.3" → "1.2.4"，
// 有预发布版本时仅去掉预发布标识，如 "1.2.3-beta" → "1.2.3"。
func IncPatch(v string) (string, error) {
	sv, err := Parse(v)
	if err != nil {
		return "", err
	}
	return sv.IncPatch().String(), nil
}

// SetPrerelease 设置预发布版本标识，传空串表示清除预发布信息。
func SetPrerelease(v, prerelease string) (string, error) {
	sv, err := Parse(v)
	if err != nil {
		return "", err
	}
	nv, err := sv.SetPrerelease(prerelease)
	if err != nil {
		return "", fmt.Errorf("设置预发布版本失败：%w，版本：%s，预发布：%s", err, v, prerelease)
	}
	return nv.String(), nil
}

// SetMetadata 设置构建元数据，传空串表示清除元数据。
func SetMetadata(v, metadata string) (string, error) {
	sv, err := Parse(v)
	if err != nil {
		return "", err
	}
	nv, err := sv.SetMetadata(metadata)
	if err != nil {
		return "", fmt.Errorf("设置构建元数据失败：%w，版本：%s，元数据：%s", err, v, metadata)
	}
	return nv.String(), nil
}

// Match 判断版本号是否满足约束表达式。
// 支持逗号分隔多条件与 AND（空格），如 ">=1.2.0, <2.0.0"、"^1.2.3"、"~1.2"、"1.x"、">1.0, <2.0 || >=3.0"。
// 不推荐在生产中用 Version 同时又作为约束（constraint 使用独立语法）。
func Match(v, constraint string) (bool, error) {
	sv, cs, err := parseVersionConstraint(v, constraint)
	if err != nil {
		return false, err
	}
	return cs.Check(sv), nil
}

// Validate 校验版本号是否满足约束，不满足时返回具体原因；满足时返回 nil。
func Validate(v, constraint string) error {
	sv, cs, err := parseVersionConstraint(v, constraint)
	if err != nil {
		return err
	}
	ok, errs := cs.Validate(sv)
	if ok {
		return nil
	}
	msgs := make([]string, 0, len(errs))
	for _, e := range errs {
		msgs = append(msgs, e.Error())
	}
	return fmt.Errorf("版本号 %s 不满足约束 %s：%s", v, constraint, strings.Join(msgs, "; "))
}

// Sort 对版本号列表排序，asc 为 true 时升序，false 时降序，返回规范化后的版本列表。
func Sort(vs []string, asc bool) ([]string, error) {
	versions := make([]*semver.Version, 0, len(vs))
	for _, v := range vs {
		sv, err := Parse(v)
		if err != nil {
			return nil, err
		}
		versions = append(versions, sv)
	}

	collection := semver.Collection(versions)
	if asc {
		sort.Sort(collection)
	} else {
		sort.Sort(sort.Reverse(collection))
	}

	retList := make([]string, 0, len(versions))
	for _, sv := range versions {
		retList = append(retList, sv.String())
	}
	return retList, nil
}

// Max 返回版本号列表中的最大版本。
func Max(vs ...string) (string, error) {
	if len(vs) == 0 {
		return "", fmt.Errorf("版本号列表不能为空")
	}
	maxVersion, err := Parse(vs[0])
	if err != nil {
		return "", err
	}
	for _, v := range vs[1:] {
		sv, subErr := Parse(v)
		if subErr != nil {
			return "", subErr
		}
		if sv.GreaterThan(maxVersion) {
			maxVersion = sv
		}
	}
	return maxVersion.String(), nil
}

// Min 返回版本号列表中的最小版本。
func Min(vs ...string) (string, error) {
	if len(vs) == 0 {
		return "", fmt.Errorf("版本号列表不能为空")
	}
	minVersion, err := Parse(vs[0])
	if err != nil {
		return "", err
	}
	for _, v := range vs[1:] {
		sv, subErr := Parse(v)
		if subErr != nil {
			return "", subErr
		}
		if sv.LessThan(minVersion) {
			minVersion = sv
		}
	}
	return minVersion.String(), nil
}

// parsePair 解析两个待比较的版本号
func parsePair(a, b string) (*semver.Version, *semver.Version, error) {
	va, err := Parse(a)
	if err != nil {
		return nil, nil, err
	}
	vb, err := Parse(b)
	if err != nil {
		return nil, nil, err
	}
	return va, vb, nil
}

// parseVersionConstraint 解析版本号与约束表达式
func parseVersionConstraint(v, constraint string) (*semver.Version, *semver.Constraints, error) {
	sv, err := Parse(v)
	if err != nil {
		return nil, nil, err
	}
	cs, err := semver.NewConstraint(constraint)
	if err != nil {
		return nil, nil, fmt.Errorf("约束表达式不合法：%w，约束：%s", err, constraint)
	}
	return sv, cs, nil
}
