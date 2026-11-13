package handler
import("errors";"gamematch/store";"net/http";"strconv")
func(a *API) ReviewVoteIDs(w http.ResponseWriter,r *http.Request){
 id,e:=strconv.ParseInt(r.PathValue("id"),10,64);if e!=nil||id<=0{fail(w,400,"รหัสเกมไม่ถูกต้อง");return}
 ids,e:=a.S.ReviewVoteIDs(r.Context(),id,current(r).ID);if e!=nil{dbError(w,e);return};send(w,200,ids)
}
func(a *API) ReviewVote(w http.ResponseWriter,r *http.Request){
 game,e:=strconv.ParseInt(r.PathValue("id"),10,64);id,e2:=strconv.ParseInt(r.PathValue("reviewId"),10,64)
 if e!=nil||e2!=nil||game<=0||id<=0{fail(w,400,"รหัสรีวิวไม่ถูกต้อง");return}
 add:=r.Method=="PUT"
 count,e:=a.S.SetReviewVote(r.Context(),game,id,current(r).ID,add)
 if errors.Is(e,store.ErrOwnReview){fail(w,400,"ไม่สามารถโหวตรีวิวของตัวเอง");return}
 if e!=nil{dbError(w,e);return}
 send(w,200,map[string]any{"helpfulCount":count,"voted":add})
}
