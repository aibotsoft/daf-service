package api

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var betResponse = `
// -->
</script>
<img src="Images/ALog/dafabet-mobile-logo.jpg" border="0" /><br>
<b>ID: 109031422975</b><br>
Bet Success, <a href="BetListRunning.aspx">Please Check Bet List</a><br>
&nbsp;
        <a href="underover/OddsNew.aspx">Back</a>&nbsp;
        <a href="Menu.aspx?m_t=0">Menu</a><br>
</form></body></html>
`

func Test_parseBetId(t *testing.T) {
	id, err := parseBetId(betResponse)
	if assert.NoError(t, err) {
		assert.NotEmpty(t, id)
		t.Log(id)
	}
}
