wrk.method="GET"
wrk.headers["Content-Type"] = "application/json"
-- User-Agent Authorization 两个都需要修改
wrk.headers["User-Agent"] = "PostmanRuntime/7.53.0"
wrk.headers["Authorization"]="Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3ODA0NzM2MDgsIlVpZCI6MSwiVXNlckFnZW50IjoiUG9zdG1hblJ1bnRpbWUvNy41My4wIn0.tROiuXp11kZ3YXFBnbIi6JbQ0f3URWVGsPJ1_d8o1ixat7qixvuLWcY0VJlbdptsX3dXUcz-zm-UCWBBXkeS8A"