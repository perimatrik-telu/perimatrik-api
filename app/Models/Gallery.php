<?php
namespace App\Models;

use Illuminate\Database\Eloquent\Concerns\HasUuids;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Support\Str; // Add this line

class Gallery extends Model
{
    use HasFactory, HasUuids;

    protected $guarded = [];

    /**
     * The accessors to append to the model's array form.
     *
     * @var array
     */
    protected $appends = [
        'image_url',
    ];

    /**
     * Get the gallery's image url.
     *
     * @return string
     */
    public function getImageUrlAttribute(): string
    {
        if (empty($this->image)) {
            return ''; // Return empty string if image is null or empty
        }

        if (Str::contains($this->image, '/storage/')) {
            $imagePath = Str::after($this->image, '/storage/');
            return route('image.show', ['path' => $imagePath]);
        }

        // If it's not a /storage/ path, assume it's already a full URL or relative path
        // and return it directly. Ensure it's treated as a string.
        return (string) $this->image;
    }

    /**
     * Get the user that owns the gallery.
     */
    public function user(): BelongsTo
    {
        return $this->belongsTo(User::class);
    }
}
