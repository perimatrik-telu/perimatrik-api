<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ $title }}</title>
    
    <meta property="og:title" content="{{ $title }}">
    <meta property="og:description" content="{{ $description }}">
    <meta property="og:image" content="{{ $image }}">
    <meta property="og:url" content="{{ $url }}">
    <meta property="og:type" content="article">
    
    <meta name="twitter:card" content="summary_large_image">
    <meta name="twitter:title" content="{{ $title }}">
    <meta name="twitter:description" content="{{ $description }}">
    <meta name="twitter:image" content="{{ $image }}">
    
    <meta http-equiv="refresh" content="0; url={{ $url }}">
    <script>
        window.location.replace("{!! $url !!}");
    </script>
</head>
<body>
    <p>Mengarahkan halaman... Jika tidak dialihkan secara otomatis, <a href="{{ $url }}">klik di sini</a>.</p>
</body>
</html>
