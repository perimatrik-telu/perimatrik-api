<?php
namespace App\Helpers;

use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Str;

class AttachmentHelper
{
    public static function saveImageMedia($request, $mediaName, $path)
    {
        $image     = $request->file($mediaName);
        $extension = $image->getClientOriginalExtension();
        $filename  = Str::random(30) . '.' . $extension;

        Storage::disk("public")->putFileAs($path, $image, $filename);

        return $filename;
    }

    public static function deleteMedia($path, $filename)
    {
        $filePath = $path . "/" . $filename;

        if (Storage::disk('public')->exists($filePath)) {
            return Storage::disk('public')->delete($filePath);
        }

        return true;
    }
}
