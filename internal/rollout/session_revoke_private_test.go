package rollout

import ("testing"; "time")

func TestSessionRevocationIsIdempotent(t *testing.T) {
 s:=NewStore(); now:=time.Now(); if err:=s.PutSession(Session{Token:"tok",TenantID:"t",UserID:"u",ExpiresAt:now.Add(time.Hour)}); err!=nil {t.Fatal(err)}
 if !s.RevokeSession("tok",now) {t.Fatal("first revoke failed")}
 if s.RevokeSession("tok",now.Add(time.Minute)) {t.Fatal("duplicate revoke reported success")}
}
