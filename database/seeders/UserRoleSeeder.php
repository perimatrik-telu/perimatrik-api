<?php

namespace Database\Seeders;

use App\Models\User;
use Illuminate\Database\Console\Seeds\WithoutModelEvents;
use Illuminate\Database\Seeder;
use Illuminate\Support\Facades\Hash;

class UserRoleSeeder extends Seeder
{
    /**
     * Run the database seeds.
     */
    public function run(): void
    {
        User::create([
            'name'     => 'Admin User',
            'email'    => 'admin@example.com',
            'password' => Hash::make('password'),
            'role'     => 'admin',
        ]);

        User::create([
            'name'     => 'Member User',
            'email'    => 'member@example.com',
            'password' => Hash::make('password'),
            'role'     => 'member',
        ]);

        User::create([
            'name'     => 'Officer User',
            'email'    => 'officer@example.com',
            'password' => Hash::make('password'),
            'role'     => 'officer',
        ]);
    }
}
