# OceanWay 部署运行手册

本文件来自旧 OceanWay 部署流程的最小迁移版，用于 `v0.1.137` clean 重建仓库。部署前先读本文件，不要凭旧记忆操作。

## 当前结论

- `ocean-way.top` 是完整服务入口，包含登录、控制台、后台和 `/v1` API。
- `oceanway.site` 是面向非技术用户的引流站点，只允许 `/`、`/internal-home`、`/docs`、`/agents`，不开放登录、后台、`/api/*` 或 `/v1/*`。
- 主生产是 `la-vps2` 单主机架构：Caddy、`sub2api`、PostgreSQL、Redis 都在 `la-vps2`。
- `la-vps` 只作为迁移期兼容入口，保留旧 IP `/v1` 转发。
- `ali-vps` 不再部署新业务代码，只保留旧数据和备份观察。
- 镜像 tag 是部署标记，不是应用版本号。不要使用 `--build-arg VERSION=${TAG}`。

## 拓扑

| 节点 | 地址 | 职责 |
| --- | --- | --- |
| `la-vps2` | `156.238.231.111` | 主生产节点；运行 Caddy、应用、PostgreSQL、Redis |
| `la-vps` | `64.188.30.215` | 迁移期兼容节点；旧 IP `/v1` 入口转发到 `la-vps2` |
| `ali-vps` | `43.112.125.18` | 旧生产节点；应用保持停止，仅保留数据库/备份观察 |

`la-vps2` 关键路径：

```text
/opt/sub2api/docker-compose.yml
/opt/sub2api/source
/opt/sub2api/data
/opt/sub2api/postgres_data
/opt/sub2api/redis_data
/root/sub2api-backups
/etc/caddy/Caddyfile
```

## 部署前检查

本地先确认代码范围：

```bash
git status --short --branch
git log --oneline -5 --decorate
```

推荐只部署已提交代码。`git archive HEAD` 不包含未提交改动。

推荐本地验证：

```bash
corepack pnpm --dir frontend run typecheck
corepack pnpm --dir frontend run build
(cd backend && go test ./...)
```

如果本机 Go 版本低于 `backend/go.mod`，不要改生产仓库版本；只在本地临时验证并恢复。

## 应用部署

### 1. 生成部署变量

```bash
TAG=oceanway-$(date +%Y%m%d-%H%M)
COMMIT=$(git rev-parse --short HEAD)
echo "TAG=${TAG}"
echo "COMMIT=${COMMIT}"
cat backend/cmd/server/VERSION
```

应用版本来自 `backend/cmd/server/VERSION`，页面不应显示 `voceanway-...`。

### 2A. 本地有 Docker

`la-vps2` 是 `linux/amd64`。Mac 本地构建必须指定平台：

```bash
REMOTE_PLATFORM=linux/amd64
docker build --platform "${REMOTE_PLATFORM}" -t sub2api:${TAG} --build-arg COMMIT=${COMMIT} .
docker image inspect sub2api:${TAG} --format 'OS={{.Os}} Architecture={{.Architecture}}'
docker run --rm --platform "${REMOTE_PLATFORM}" sub2api:${TAG} /app/sub2api --version
docker save sub2api:${TAG} | gzip > /tmp/sub2api-${TAG}.tar.gz
scp /tmp/sub2api-${TAG}.tar.gz la-vps2:/root/
```

在 `la-vps2` 加载镜像：

```bash
ssh la-vps2 "TAG='${TAG}' bash -s" <<'EOF'
set -euo pipefail
gunzip -c "/root/sub2api-${TAG}.tar.gz" | docker load
docker image inspect "sub2api:${TAG}" --format 'OS={{.Os}} Architecture={{.Architecture}}'
docker run --rm "sub2api:${TAG}" /app/sub2api --version
EOF
```

### 2B. 本机没有 Docker

使用 `git archive HEAD` 上传源码到 `la-vps2` 远端构建：

```bash
TAG=oceanway-$(date +%Y%m%d-%H%M)
COMMIT=$(git rev-parse --short HEAD)
ARCHIVE=/tmp/sub2api-source-${TAG}-${COMMIT}.tar.gz
REMOTE_ARCHIVE=/root/$(basename "${ARCHIVE}")
git archive --format=tar HEAD | gzip > "${ARCHIVE}"
scp "${ARCHIVE}" la-vps2:/root/
```

如果必须部署未提交工作区，先明确风险，且在 Mac 上必须过滤 AppleDouble 和 `.DS_Store`：

```bash
COPYFILE_DISABLE=1 git ls-files -coz --exclude-standard \
  | COPYFILE_DISABLE=1 tar --null -czf "${ARCHIVE}" --files-from -

if tar -tzf "${ARCHIVE}" | grep -E '(^|/)(\._|\.DS_Store$|__MACOSX/)'; then
  echo "ERROR: macOS metadata files found in archive; abort deploy." >&2
  exit 1
fi
```

远端备份旧源码并构建：

```bash
ssh la-vps2 "TAG='${TAG}' COMMIT='${COMMIT}' ARCHIVE='${REMOTE_ARCHIVE}' bash -s" <<'EOF'
set -euo pipefail
BACKUP_DIR=/root/sub2api-backups/pre-deploy-${TAG}-$(date +%Y%m%d-%H%M%S)

mkdir -p "${BACKUP_DIR}"
cp -a /opt/sub2api/docker-compose.yml "${BACKUP_DIR}/docker-compose.yml"
cp -a /opt/sub2api/source "${BACKUP_DIR}/source"

rm -rf /opt/sub2api/source
mkdir -p /opt/sub2api/source
tar -xzf "${ARCHIVE}" -C /opt/sub2api/source

if find /opt/sub2api/source \( -name '._*' -o -name '.DS_Store' -o -name '__MACOSX' \) -print -quit | grep -q .; then
  echo "ERROR: macOS metadata files found after extraction; abort deploy." >&2
  exit 1
fi

cd /opt/sub2api/source
cat backend/cmd/server/VERSION
docker build -t sub2api:${TAG} --build-arg COMMIT=${COMMIT} .
docker run --rm sub2api:${TAG} /app/sub2api --version
EOF
```

### 3. 切换线上镜像

```bash
ssh la-vps2 "TAG='${TAG}' COMMIT='${COMMIT}' bash -s" <<'EOF'
set -euo pipefail
cd /opt/sub2api

cp docker-compose.yml docker-compose.yml.bak-before-${TAG}-$(date +%Y%m%d-%H%M%S)
perl -0pi -e "s|image: sub2api:[^\n]+|image: sub2api:${TAG}|; s|COMMIT: [^\n]+|COMMIT: ${COMMIT}|" docker-compose.yml

docker compose up -d --no-build
docker compose ps
docker inspect sub2api --format 'Image={{.Config.Image}} Status={{.State.Status}} Health={{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}'
docker exec sub2api /app/sub2api --version
docker logs --tail=120 sub2api
EOF
```

必须在 `/opt/sub2api` 执行 compose。不要在 `la-vps` 或 `ali-vps` 切应用镜像。

## 部署后验证

```bash
for url in \
  https://ocean-way.top/health \
  https://ocean-way.top/home \
  https://ocean-way.top/docs \
  https://ocean-way.top/agents \
  https://ocean-way.top/login \
  https://ocean-way.top/v1/models \
  https://oceanway.site/internal-home \
  https://oceanway.site/docs \
  https://oceanway.site/agents \
  https://oceanway.site/login \
  https://oceanway.site/v1/models \
  http://64.188.30.215:8080/v1/models \
  http://64.188.30.215:8080/home; do
  printf '%s -> ' "$url"
  curl -k -sS -o /dev/null -w '%{http_code} %{remote_ip} %{redirect_url}\n' "$url" || true
done
```

预期：

| URL | 预期 |
| --- | --- |
| `https://ocean-way.top/health` | `200` |
| `https://ocean-way.top/home` | `200` |
| `https://ocean-way.top/docs` | `200` |
| `https://ocean-way.top/agents` | `200` |
| `https://ocean-way.top/login` | `200` |
| `https://ocean-way.top/v1/models` | 未带 Key 通常 `401` |
| `https://oceanway.site/internal-home` | `200` |
| `https://oceanway.site/docs` | `200` |
| `https://oceanway.site/agents` | `200` |
| `https://oceanway.site/login` | `302` 到 `/internal-home` |
| `https://oceanway.site/v1/models` | `404` |
| `http://64.188.30.215:8080/v1/models` | 未带 Key 通常 `401`，仅迁移期保留 |
| `http://64.188.30.215:8080/home` | `404` |

版本验证：

```bash
curl -k -sS https://ocean-way.top/api/v1/settings/public
```

`data.version` 应等于 `backend/cmd/server/VERSION`。

## Caddy 边界

只有修改入口分流或域名边界时才操作 Caddy。

`la-vps2` 主入口应只反代本机 `127.0.0.1:8080`：

```caddy
(oceanway_headers) {
	header {
		Strict-Transport-Security "max-age=31536000; includeSubDomains; preload"
		X-Content-Type-Options "nosniff"
		Referrer-Policy "strict-origin-when-cross-origin"
	}
}

(oceanway_app_proxy) {
	reverse_proxy 127.0.0.1:8080 {
		flush_interval -1
		header_up -If-None-Match
		header_up -If-Modified-Since
		header_down -ETag
	}
}

ocean-way.top, www.ocean-way.top {
	import oceanway_app_proxy
	import oceanway_headers
}

oceanway.site, www.oceanway.site {
	@blocked_api path /api /api/* /v1 /v1/*
	respond @blocked_api 404

	@blocked_app path /home /login /register /email-verify /forgot-password /reset-password /dashboard /dashboard/* /admin /admin/* /keys /keys/* /usage /usage/* /subscriptions /subscriptions/* /redeem /redeem/* /purchase /purchase/* /payment /payment/* /auth /auth/* /setup
	redir @blocked_app /internal-home 302

	import oceanway_app_proxy
	import oceanway_headers
}
```

修改后：

```bash
ssh la-vps2 'sudo caddy validate --config /etc/caddy/Caddyfile && sudo systemctl reload caddy'
```

`la-vps` 旧 IP 兼容入口只允许 `/v1`：

```caddy
(oceanway_top_legacy_ip_origin) {
	reverse_proxy https://156.238.231.111 {
		flush_interval -1
		header_up Host ocean-way.top
		transport http {
			tls_server_name ocean-way.top
		}
	}
}

http://64.188.30.215:8080 {
	bind 64.188.30.215

	@legacy_v1 path /v1 /v1/*
	handle @legacy_v1 {
		import oceanway_top_legacy_ip_origin
	}

	handle {
		respond 404
	}
}
```

如果 `la-vps` Caddy 全局是 `admin off`，验证通过后用 `systemctl restart caddy`，不要用 reload。

## 后台配置

部署后后台设置应保持：

```text
internal_home_domains = oceanway.site
site_name = OceanWay AI
api_base_url = https://ocean-way.top/v1
```

不要把 `api_base_url` 写成 `https://oceanway.site/v1`，因为 `oceanway.site` 不提供 API 服务。

## Key 路由说明

当前 clean 重建已经在业务代码里实现：

- 新注册用户自动获得默认 Key。
- 用户侧新建 Key 默认不绑定分组。
- 未绑定分组的 Key 请求时按“订阅优先、余额兜底”选择可用分组。

旧仓库文档里的 `backend/cmd/migrate-key-routing` 是历史迁移脚本，本轮 clean 重建仓库没有迁移该命令。不要在 `la-vps2` 执行不存在的历史脚本；如需批量修复旧线上数据，应另开一次性脚本任务并先做数据库备份。

## 备份

在 `la-vps2` 执行：

```bash
ssh la-vps2 'set -euo pipefail
TS=$(date +%Y%m%d-%H%M%S)
DIR=/root/sub2api-backups/manual-${TS}
mkdir -p "$DIR"
cd /opt/sub2api
sudo docker compose exec -T postgres pg_dump -U sub2api -d sub2api -Fc --no-owner --no-acl > "$DIR/sub2api.dump"
tar --warning=no-file-changed --exclude=".DS_Store" --exclude="._*" --exclude="__MACOSX" -C /opt/sub2api -czf "$DIR/sub2api-data.tar.gz" data || test $? -eq 1
cp .env docker-compose.yml "$DIR/"
sha256sum "$DIR"/* > "$DIR/SHA256SUMS"
ls -lh "$DIR"
'
```

## 回滚

找上一版镜像：

```bash
ssh la-vps2 "docker images --format '{{.Repository}}:{{.Tag}} {{.CreatedSince}}' | grep '^sub2api:'"
```

切回上一版：

```bash
ssh la-vps2 "ROLLBACK_TAG='oceanway-上一版' bash -s" <<'EOF'
set -euo pipefail
cd /opt/sub2api
cp docker-compose.yml docker-compose.yml.bak-rollback-$(date +%Y%m%d-%H%M%S)
perl -0pi -e "s|image: sub2api:[^\n]+|image: sub2api:${ROLLBACK_TAG}|" docker-compose.yml
docker compose up -d --no-build
docker compose ps
docker inspect sub2api --format 'Image={{.Config.Image}} Status={{.State.Status}} Health={{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}'
EOF
```

如果改过 Caddy，恢复对应服务器的 Caddy 备份后执行：

```bash
caddy validate --config /etc/caddy/Caddyfile
systemctl reload caddy
```

`la-vps` 若 `admin off`，使用 `systemctl restart caddy`。

## 常见错误

1. 在 `ali-vps` 部署新应用镜像。
2. 在 `la-vps` 切应用镜像，以为主站会变化。
3. 把镜像 tag 当应用版本。
4. 使用 `latest` 做生产版本。
5. 让 `oceanway.site` 暴露 API、登录或后台。
6. 从 Mac 工作区直接 `tar -czf source .`，带入 `._*` 或 `.DS_Store`。
7. 旧 IP 兼容入口放开过多路径。
8. 旧 IP 转发没有固定 `Host: ocean-way.top`。
9. 重新依赖 `origin.oceanway.site`。
10. 泄露 `/opt/sub2api/.env`。
