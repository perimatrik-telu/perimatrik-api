<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Run the migrations.
     */
    public function up(): void
    {
        Schema::create('articles', function (Blueprint $table) {
            $table->uuid('id')->primary()->unique();
            $table->string('slug');
            $table->binary('image');
            $table->string('title');
            $table->longText('body');
            $table->unsignedInteger('views')->default(0);
            $table->enum('status', ['draft', 'active', 'inactive'])->default('draft');
            $table->foreignUuid('user_id')->nullable()->constrained('users')->onDelete('cascade')->nullOnDelete();
            $table->timestamps();
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists('articles');
    }
};
