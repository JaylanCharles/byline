# 基础镜像
FROM ubuntu:22.04
# 把编译后的打包进来这个镜像，放到工作目录 /app 这个随便换，看个人喜好和公司规定
COPY byline /app/byline
WORKDIR /app
# CMD 也是执行命令，和 ENTRYPOINT 没什么区别，最佳实践是使用 ENTRYPOINT
ENTRYPOINT ["/app/byline"]