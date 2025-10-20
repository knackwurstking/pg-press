# Troubleshooting

If the database get removed or replaced, the server is completely broken.
For example, i can't load the tools sections (press, all tools), or the profile cookies are missing

```bash
[DEBUG] 2025/10/20 22:16:58 [Service: Users] Getting user by API key
[DEBUG] 2025/10/20 22:16:58 [Service: Cookies] Removing cookie
[ERROR] 2025/10/20 22:16:58 [Auth] Failed to remove existing cookie for user test-user from ::1: not found: cookie with ID 2c29f4a1-ab65-4d59-9024-9a53cb444680 not found
[DEBUG] 2025/10/20 22:16:58 [Auth] Creating cookie for user test-user from ::1: HTTPS=false, UserAgent='Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/26.0.1 Safari/605.1.15', Scheme='http', Path='/pg-press'
[DEBUG] 2025/10/20 22:16:58 [Auth] Cookie set successfully: Name=pgpress-api-key, Secure=false, SameSite=2, Path=/pg-press, Expires=2026-04-24 22:16:58.46558 +0200 CEST m=+16070433.267386293
[DEBUG] 2025/10/20 22:16:58 [Service: Cookies] Adding cookie
[DEBUG] 2025/10/20 22:16:58 [Auth] Session cookie stored in database for user test-user
[INFO ] 2025/10/20 22:16:58 [Auth] Successful login for user from ::1
↩️  2025/10/20 22:16:58 [Server] 303 GET     /pg-press/login?api-key=pgp_hh7V0hQu17e393K2YLH5XNPZkh826441_399563034 (::1) 928.458µs
[DEBUG] 2025/10/20 22:16:58 [Middleware: Cookie Validation] Found cookie for request from ::1: expires=0001-01-01 00:00:00 +0000 UTC, secure=false
[DEBUG] 2025/10/20 22:16:58 [Service: Cookies] Getting cookie by value
[DEBUG] 2025/10/20 22:16:58 [Middleware: Cookie Validation] Cookie found in database for request from ::1: lastLogin=1760991418465, userAgent='Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/26.0.1 Safari/605.1.15'
[DEBUG] 2025/10/20 22:16:58 [Middleware: Cookie Validation] Cookie is valid (not expired) for request from ::1
[DEBUG] 2025/10/20 22:16:58 [Service: Users] Getting user by API key
[DEBUG] 2025/10/20 22:16:58 [Middleware: Cookie Validation] User validated from cookie for request from ::1: user=test-user
[DEBUG] 2025/10/20 22:16:58 [Middleware: Cookie Validation] Updating cookie for page visit by user test-user from ::1: path=/pg-press/profile
[DEBUG] 2025/10/20 22:16:58 [Service: Cookies] Updating cookie
[DEBUG] 2025/10/20 22:16:58 [Middleware: Cookie Validation] Cookie successfully updated for user test-user from ::1
[DEBUG] 2025/10/20 22:16:58 [Middleware: Cookie Validation] Cookie validation successful for user test-user from ::1
[DEBUG] 2025/10/20 22:16:58 [Profile] Rendering profile page for user test-user
✅ 2025/10/20 22:16:58 [Server] 200 GET     /pg-press/profile (::1) 791.5µs User{ID: 1, Name: test-user [has API key]}
✅ 2025/10/20 22:16:58 [Server] 200 GET     /pg-press/login (::1) 116.583µs
✅ 2025/10/20 22:16:58 [Server] 200 GET     /pg-press/favicon.ico?v=1760991386 (::1) 45.792µs
✅ 2025/10/20 22:16:58 [Server] 200 GET     /pg-press/icon.png?v=1760991386 (::1) 85.416µs
```
