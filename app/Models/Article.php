<?php
namespace App\Models;

use Illuminate\Database\Eloquent\Concerns\HasUuids;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Support\Carbon;

class Article extends Model
{
    use HasFactory, HasUuids;

    public $guarded = [];

    /**
     * The accessors to append to the model's array form.
     *
     * @var array
     */
    protected $appends = [
        'author',
        'postingAt',
        'tags',
    ];

    /**
     * The attributes that should be cast.
     *
     * @var array<string, string>
     */
    protected function casts(): array
    {
        return [
            'published_at' => 'datetime',
        ];
    }

    /**
     * Get the route key for the model.
     */
    public function getRouteKeyName(): string
    {
        return 'slug';
    }

    /**
     * Get the user that owns the Article
     *
     * @return \Illuminate\Database\Eloquent\Relations\BelongsTo
     */
    public function user(): BelongsTo
    {
        return $this->belongsTo(\App\Models\User::class);
    }

    /**
     * Get the article's author.
     *
     * @return string|null
     */
    public function getAuthorAttribute(): ?string
    {
        return $this->user ? $this->user->name : null;
    }

    /**
     * Get the article's posting date.
     *
     * @return string
     */
    public function getPostingAtAttribute(): string
    {
        return Carbon::parse($this->published_at ?? $this->created_at)->toIso8601String();
    }

    /**
     * Get the article's tags.
     *
     * @return array
     */
    public function getTagsAttribute(): array
    {
        return [];
    }
}