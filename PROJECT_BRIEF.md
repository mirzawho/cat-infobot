# PROJECT BRIEF — InfoBot

## هدف

ربات تلگرام **InfoBot** که هر پیامی را که برایش ارسال یا Forward شود دریافت می‌کند و
اطلاعات فنی و جزئیات قابل‌دسترسی آن را به فارسی نمایش می‌دهد. بدون دیتابیس، بدون تنظیمات
کاربر و بدون پنل مدیریت.

## فناوری

| مورد | انتخاب |
| --- | --- |
| زبان | Go 1.27 (فقط Standard Library) |
| Telegram | Bot API (HTTPS JSON) با getUpdates / Long Polling |
| اجرا | Docker، Docker Compose، Kubernetes |
| CI/CD | GitHub Actions، انتشار Image در GHCR |

## معماری

```
main.go              راه‌اندازی، سیگنال‌ها، قفل تک‌نمونه، healthcheck
internal/telegram/   کلاینت Bot API، Poller، مدل Update/Message، قفل فایل
internal/bot/        /start ، خروجی JSON/Markdown، تقسیم خروجی، ارسال
```

منطق ساخت خروجی (`internal/bot/render.go`) کاملاً از شبکه جدا است و با JSON نمونه تست می‌شود.

## خروجی

خروجی یک پاسخ MarkdownV2 است که همان پیام را به‌صورت JSON (indent شده) داخل code block
برمی‌گرداند و به پیام کاربر Reply می‌شود. منطق ساخت آن (`internal/bot/render.go`) از شبکه
جدا است و با JSON نمونه تست می‌شود.

## تصمیم‌های کلیدی

- تنها `BOT_TOKEN` (و اختیاری `TELEGRAM_API_ENDPOINT`) لازم است؛ `api_id`/`api_hash` لازم نیست.
- خروجی MarkdownV2 است: فقط heading با Markdown نوشته می‌شود و محتوای JSON داخل code block
  escape می‌شود، بنابراین هیچ کاراکتر پیام خطا نمی‌سازد.
- فقط یک نمونه هم‌زمان (قفل فایل با flock)، Shutdown صحیح با SIGINT/SIGTERM.
- خطاهای موقت شبکه و `429` باعث Crash نمی‌شوند؛ فقط خطای 401 (توکن نامعتبر) توقف می‌دهد.
- هیچ فیلدی حذف یا حدس زده نمی‌شود؛ کل Update همان‌طور که رسیده نمایش داده می‌شود.

## فایل‌ها

`main.go`، `internal/**`، `*_test.go`، `Dockerfile`، `docker-compose.yml`، `k8s.yaml`،
`.github/workflows/deploy.yml`، `.env.example`، `.dockerignore`، `.gitignore`، `Makefile`،
`PROJECT_BRIEF.md`، `README.md`، `go.mod`.
