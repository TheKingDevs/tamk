# ProGuard Configuration for T.A.M.K Guardian Security
# Usage: proguard @proguard-android.pro

# Input/Output
-injars obj/
-outjars obfuscated/
-libraryjars /root/.tamk/development/sdk/android.jar

# Keep attributes for debugging (optional)
-keepattributes SourceFile,LineNumberTable

# Aggressive obfuscation
-repackageclasses ''
-allowaccessmodification
-overloadaggressively

# Remove logging
-assumenosideeffects class android.util.Log {
    public static int v(...);
    public static int d(...);
    public static int i(...);
    public static int w(...);
    public static int e(...);
}

# Keep Android components
-keep public class * extends android.app.Activity
-keep public class * extends android.app.Application
-keep public class * extends android.app.Service
-keep public class * extends android.content.BroadcastReceiver
-keep public class * extends android.content.ContentProvider

# Keep annotations
-keepattributes *Annotation*

# Keep security classes (integrity check)
-keep class **.security.GuardianBridge { *; }
-keep class **.security.RASPSecurityModule { *; }
-keep class **.security.CertificatePinner { *; }
-keep class **.security.IntegrityVerifier { *; }
-keep class **.security.StringObfuscator { *; }

# Obfuscate everything else aggressively
-repackageclasses ''
-allowaccessmodification
-optimizationpasses 5
