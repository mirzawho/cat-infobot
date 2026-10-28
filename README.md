# InfoBot — Telegram Message Inspector

ربات تلگرامی که به هر پیامی **Reply** می‌کند و همان پیام را به‌صورت JSON (داخل یک code block
با فرمت Markdown) برمی‌گرداند؛ یعنی همهٔ فیلدهایی که تلگرام برای آن پیام فرستاده، بدون حذف یا
حدس. قالب خروجی:

````text
```json
{
  "message_id": 3,
  "from": {"id": 1780243209, "first_name": "Mirza", "username": "mirza"},
  "chat": {"id": 1780243209, "type": "private"},
  "text": "سلام"
  ...
}
```
````

فقط همین: نه دیتابیس، نه پنل مدیریت، نه تنظیمات کاربر.

## چه چیزی نیاز دارید

فقط **توکن ربات** (از [@BotFather](https://t.me/BotFather)) و به‌صورت اختیاری آدرس یک Bot API
سرور دیگر (`TELEGRAM_API_ENDPOINT`). هیچ `api_id` / `api_hash` لازم نیست.

Image پروژه با هر Push به `main` روی GitHub Actions ساخته و در `ghcr.io/mirzawho/cat-infobot` منتشر می‌شود.

## دریافت توکن

1. به `@BotFather` پیام بدهید و `/newbot` را بزنید.
2. نام و username ربات را وارد کنید.
3. توکن قالب `123456789:AA...` را در `.env` قرار دهید. هرگز آن را Commit نکنید.

## اجرای محلی

```bash
cp .env.example .env      # BOT_TOKEN را پر کنید
make check                # gofmt + vet + test
make run                  # اجرا از سورس (محتوای .env را می‌خواند)
```

یا مستقیم:

```bash
DATA_DIR=/tmp/infobot-data go build -o infobot . && ./infobot
```

در تلگرام به ربات پیام بدهید یا `/start` را بزنید.

## اجرای Docker Compose

```bash
cp .env.example .env        # BOT_TOKEN
make compose-up             # یا: docker compose up -d --build
make compose-logs
make compose-down
```

پورت منتشر نمی‌شود چون ربات فقط Long Polling می‌کند و دامنه عمومی نمی‌خواهد.

### Docker خام

```bash
make docker IMAGE=infobot TAG=local
make docker-run IMAGE=infobot TAG=local
docker inspect --format '{{.State.Health.Status}}' infobot
```

## استقرار روی Kubernetes

1. Secret بسازید (توکن واقعی را در YAML نگذارید):

```bash
kubectl create namespace infobot
kubectl -n infobot create secret generic infobot-secret \
  --from-literal=BOT_TOKEN='123456789:AA...'
```

2. Image را با تگ commit مورد نظر در `k8s.yaml` تنظیم کنید و اعمال کنید:

```bash
kubectl apply -f k8s.yaml
kubectl -n infobot rollout status deployment/infobot
```

`k8s.yaml` شامل ConfigMap، Deployment (replica = ۱، سیاست `Recreate` تا دو Poller هم‌زمان
اجرا نشود)، Resource requests/limits، محدودیت‌های امنیتی و Volume برای فایل قفل است.
Service و Ingress لازم نیست.

## انتشار Image در GHCR

Workflow [`.github/workflows/deploy.yml`](.github/workflows/deploy.yml):

- روی هر Push به شاخه پیش‌فرض و هر Pull Request: `gofmt`، `go vet`، `go test -race`، `go build`
  و Build تصویر Docker
- روی Push به شاخه پیش‌فرض (اگر `VERSION` تگ خورده یا push عادی باشد): انتشار
  `ghcr.io/mirzawho/cat-infobot:<commit-sha>` و `:latest` با `GITHUB_TOKEN`

تگ زدن برای انتشار: `make release VERSION=v1.0.0`

برای Deploy خودکار: Variable `K8S_DEPLOY_ENABLED=true` و Secret `KUBE_CONFIG` (base64
kubeconfig) را تنظیم کنید؛ Variable اختیاری `K8S_NAMESPACE` (پیش‌فرض `infobot`).
بدون آن‌ها Build و انتشار Image همچنان انجام می‌شود و مرحله Deploy رد می‌شود.

## Makefile

```bash
make help          # همه دستورها
make build         # ساخت bin/infobot
make run           # اجرا با .env
make check         # gofmt + vet + test
make test-race     # تست با race detector
make docker        # ساخت Image (IMAGE و TAG)
make docker-run    # اجرای Image با .env
make compose-up    # ساخت و اجرای Compose
make k8s-dry-run   # اعتبارسنجی k8s.yaml با kubectl
make release VERSION=v1.0.0   # تگ و push برای انتشار Image
```

## تنظیمات محیطی

| متغیر | الزامی | توضیح |
| --- | --- | --- |
| `BOT_TOKEN` | بله | توکن BotFather |
| `TELEGRAM_API_ENDPOINT` | خیر | آدرس Bot API (پیش‌فرض `https://api.telegram.org`) |
| `DATA_DIR` | خیر | مسیر فایل قفل تک‌نمونه (پیش‌فرض `/data`) |
| `INFOBOT_LOG_LEVEL` | خیر | `debug`، `info`، `warn`، `error` |

## محدودیت‌ها

- وقتی پیامی از کانال/گروهی Forward شده باشد، خطی مثل
  `Forwarded from channel "Name" (chat_id: -100…, source message_id: …)` بالای JSON می‌آید.
  اگر تلگرام مبدأ را مخفی کرده باشد (hidden forwarding)، به‌جای حدس‌زدن همین را صریح می‌نویسد.
- ربات فقط اطلاعاتی را نشان می‌دهد که Bot API در Update می‌فرستد. پیام Forward محافظت‌شده
  ممکن است مبدأ را نداشته باشد؛ در آن صورت خروجی صریحاً «در دسترس نیست» نشان می‌دهد.
- Entityها با نوع/آفست/طول و متن دقیق همان بازه نمایش داده می‌شوند؛ متن پیام هرگز تغییر
  نمی‌کند.
- خروجی MarkdownV2 است؛ فقط عنوان با Markdown نوشته می‌شود و JSON داخل code block escape می‌شود،
  بنابراین کاراکترهای خاص پیام هرگز خطا نمی‌سازند.
- خروجی بلندتر از ۴۰۹۶ کاراکتر روی مرز خطوط به چند پیام تقسیم می‌شود و هیچ متنی حذف نمی‌شود.
- اگر تلگرام `429 Too Many Requests` بدهد، ارسال پس از `retry_after` (تا یک دقیقه) تکرار
  می‌شود.

## عیب‌یابی

- **پاسخ‌ها به پیام کاربر Reply نمی‌شوند:** برنامه `reply_parameters` را در بدنهٔ JSON می‌فرستد.
  برای دیدن بدنهٔ دقیق درخواست و پاسخ سرور، برنامه را با `INFOBOT_DEBUG_API=1` اجرا کنید
  (توکن هرگز لاگ نمی‌شود). اگر سرور Reply را نمی‌پذیرد، گزارش
  [`SUPPORT_REPORT.md`](SUPPORT_REPORT.md) را برای پشتیبانی سرور بفرستید.
- **خطای `BOT_TOKEN is not set`:** `.env` باید در پوشه‌ای باشد که برنامه از آن اجرا می‌شود
  (یا `ENV_FILE=/path/to/.env` را تنظیم کنید)؛ متغیرهای Environment اولویت دارند.
