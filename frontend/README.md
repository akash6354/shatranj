# frontend

A new Flutter project.

The repository includes `.env.example` as a reference for the API URL.
Flutter reads `SHATRANJ_API_URL` at build time through `--dart-define`; it does
not load `.env` files automatically.

For Chrome:

```powershell
flutter run -d chrome --dart-define=SHATRANJ_API_URL=http://localhost:8080/api/v1
```

For a physical Android device, replace the host with the computer's LAN IP:

```powershell
flutter run -d 254a2537 --dart-define=SHATRANJ_API_URL=http://192.168.0.7:8080/api/v1
```
