# 📄 Templates Completos para WebApp

Código-fonte completo de todos os templates em `assets/templates/webapp/`.

---

## AndroidManifest.xml.tmpl

```xml
<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android" 
    package="{{PACKAGE}}"
    android:versionCode="1"
    android:versionName="{{VERSION}}">

    <uses-sdk android:minSdkVersion="21" android:targetSdkVersion="30" />
    <uses-permission android:name="android.permission.INTERNET" />

    <application
        android:allowBackup="true"
        android:icon="@mipmap/ic_launcher"
        android:label="{{NAME}}"
        android:theme="@style/AppTheme"
        android:usesCleartextTraffic="true"
        android:roundIcon="@mipmap/ic_launcher_round">
        
        <activity
            android:name=".MainActivity"
            android:exported="true"
            android:configChanges="orientation|screenSize|keyboardHidden"
            android:theme="@style/AppTheme.NoActionBar">
            <intent-filter>
                <action android:name="android.intent.action.MAIN" />
                <category android:name="android.intent.category.LAUNCHER" />
            </intent-filter>
        </activity>
    </application>
</manifest>
```

**Destaques:**
- `INTERNET` — essencial para WebView carregar conteúdo
- `usesCleartextTraffic` — permite HTTP (considere remover em produção)
- `configChanges` — evita recriação da activity na rotação

---

## MainActivity.kt.tmpl

```kotlin
package {{PACKAGE}}

import android.app.Activity
import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.os.Bundle
import android.webkit.WebChromeClient
import android.webkit.WebView
import android.webkit.WebViewClient
import android.webkit.WebSettings
import android.util.Log
import android.widget.Toast

class MainActivity : Activity() {

    private lateinit var webView: WebView

    // BroadcastReceiver para HMR refresh
    private val refreshReceiver = object : BroadcastReceiver() {
        override fun onReceive(context: Context?, intent: Intent?) {
            when (intent?.action) {
                "tamk.ACTION_REFRESH_ASSET" -> {
                    val path = intent.getStringExtra("path")
                    refreshAsset(path)
                }
                "tamk.ACTION_REFRESH_ALL" -> webView.reload()
            }
        }
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        webView = WebView(this)

        webView.settings.apply {
            javaScriptEnabled = true
            domStorageEnabled = true
            allowFileAccess = true
            allowContentAccess = false
            cacheMode = WebSettings.LOAD_DEFAULT
            setSupportZoom(true)
            builtInZoomControls = true
            displayZoomControls = false
        }

        webView.webViewClient = object : WebViewClient() {
            override fun onPageFinished(view: WebView?, url: String?) {
                injectHMRBridge()
            }
        }

        webView.webChromeClient = object : WebChromeClient() {
            override fun onConsoleMessage(m: android.webkit.ConsoleMessage?): Boolean {
                if (m != null) Log.d(TAG, "JS: ${m.message()}")
                return true
            }
        }

        setContentView(webView)
        webView.loadUrl("{{WEB_URL}}")
    }

    override fun onResume() {
        super.onResume()
        val filter = IntentFilter().apply {
            addAction("tamk.ACTION_REFRESH_ASSET")
            addAction("tamk.ACTION_REFRESH_ALL")
        }
        registerReceiver(refreshReceiver, filter)
    }

    override fun onPause() {
        super.onPause()
        unregisterReceiver(refreshReceiver)
    }

    override fun onBackPressed() {
        if (webView.canGoBack()) webView.goBack()
        else super.onBackPressed()
    }

    private fun injectHMRBridge() {
        // Injeta JavaScript bridge para HMR
        val bridgeJS = """
            (function() {
                if (typeof window.TAMK_DEV === 'undefined') {
                    window.TAMK_DEV = { ws: null, connected: false, hmr: { enabled: true, debug: false } };
                }
                if (typeof window.TAMK_HMR === 'undefined') {
                    window.TAMK_HMR = { _handlers: {}, accept: function(p, h) { ... }, saveState: function(k, v) { ... }, getState: function(k) { ... }, dispose: function(p) { ... } };
                }
            })();
        """.trimIndent()
        webView.evaluateJavascript(bridgeJS, null)
    }
}
```

**Destaques:**
- `injectHMRBridge()` — injeta API JS para hot reload
- `BroadcastReceiver` — recebe comandos de refresh do ADB
- `onBackPressed()` — navegação no histórico WebView
- `WebChromeClient.onConsoleMessage()` — loga console JS

---

## strings.xml.tmpl

```xml
<?xml version="1.0" encoding="utf-8"?>
<resources>
    <string name="app_name">{{NAME}}</string>
    <string name="app_version">{{VERSION}}</string>
    <string name="app_author">{{AUTHOR}}</string>
    <string name="tamk_version">{{TAMK_VERSION}}</string>
</resources>
```

---

## styles.xml.tmpl

```xml
<?xml version="1.0" encoding="utf-8"?>
<resources>
    <style name="AppTheme" parent="android:Theme.Material.Light">
        <item name="android:colorPrimary">#6200EE</item>
        <item name="android:colorPrimaryDark">#3700B3</item>
        <item name="android:colorAccent">#03DAC5</item>
    </style>
    <style name="AppTheme.NoActionBar" parent="AppTheme">
        <item name="android:windowNoTitle">true</item>
        <item name="android:windowActionBar">false</item>
    </style>
</resources>
```

---

## network_security_config.xml.tmpl

```xml
<?xml version="1.0" encoding="utf-8"?>
<network-security-config>
    <base-config cleartextTrafficPermitted="true">
        <trust-anchors>
            <certificates src="system" />
        </trust-anchors>
    </base-config>
</network-security-config>
```

---

## icon.xml.tmpl

```xml
<?xml version="1.0" encoding="utf-8"?>
<vector xmlns:android="http://schemas.android.com/apk/res/android"
    android:width="108dp" android:height="108dp"
    android:viewportWidth="108" android:viewportHeight="108">
    <path android:fillColor="#6200EE"
        android:pathData="M54,54m-40,0a40,40 0,1 1,80 0a40,40 0,1 1,-80 0"/>
    <path android:fillColor="#FFFFFF"
        android:pathData="M44,38h20v4h-20zM44,46h20v4h-20zM44,54h20v4h-20z"/>
</vector>
```

---

<!-- Templates HTML/CSS/JS estão nos arquivos .tmpl correspondentes -->

<div align="center">
  <sub>T.A.M.K v2026.3.0-HMR — Templates WebApp</sub>
</div>
