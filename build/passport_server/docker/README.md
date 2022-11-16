
构建镜像

测试版可以把配置打在镜像里，正式版单独配置，不要将配置打进镜像，具体参照对应的Dockerfile

eg：

###build：


```bash
docker build -t saas-passport-server --target=saas-passport-server -f /Users/rebuild/root/local/gitlab/saas_service/build/passport_server/docker/Dockerfile  /Users/rebuild/root/local/gitlab/saas_service/

docker tag saas-passport-server registry.xxxxxxx.com/xxx_saas/passport-test:0.01-004

docker push saas-passport-server rregistry.xxxxxxx.com/xxx_saas/passport-test:0.01-004

```

###run：
```bash
docker run -d -p 8029:8028 saas-passport-server
```