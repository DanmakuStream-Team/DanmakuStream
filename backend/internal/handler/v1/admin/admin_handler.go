package admin

import (
	"net/http"

	adminlogic "danmakustream/backend/internal/logic/admin"
	"danmakustream/backend/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func VideoListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req adminlogic.AdminVideoListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := adminlogic.NewVideoListLogic(r.Context(), svcCtx)
		resp, err := l.List(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}