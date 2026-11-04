package activity

import "testing"

// keyPrefix 中的正则元字符需要按字面匹配
func TestGetResponseRegExpQuoteMeta(t *testing.T) {
	re := getResponseRegExp("p.")
	if !re.MatchString("{{p.id1." + Arguments + ".x}}") {
		t.Fatal("should match literal prefix")
	}
	if re.MatchString("{{pXid1." + Arguments + ".x}}") {
		t.Fatal("'.' in prefix must not match any char")
	}
}
