<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Concerns\HasUuids;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;

class Member extends Model
{
    use HasFactory, HasUuids;

    protected $fillable = [
        'no_anggota',
        'nama_lengkap',
        'nama_lapangan',
        'domisili',
        'kontak',
        'avatar',
        'angkatan_id',
        'user_id',
        'instagram',
        'facebook',
        'linkedin',
        'tiktok',
        'youtube',
    ];

    public function angkatan()
    {
        return $this->belongsTo(Angkatan::class);
    }
}
