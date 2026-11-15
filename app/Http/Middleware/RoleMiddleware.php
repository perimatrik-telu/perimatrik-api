<?php

namespace App\Http\Middleware;

use App\Enums\StatusCode;
use App\Helpers\ResponseHelper;
use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;
use Illuminate\Support\Facades\Auth;

class RoleMiddleware
{
    /**
     * Handle an incoming request.
     *
     * @param  \Closure(\Illuminate\Http\Request): (\Symfony\Component\HttpFoundation\Response)  $next
     * @param  string ...$roles
     */
    public function handle(Request $request, Closure $next, string ...$roles): Response
    {
        if (!Auth::check()) {
            return ResponseHelper::error(StatusCode::UNAUTHORIZED, 'Unauthenticated.');
        }

        $user = Auth::user();

        if (!in_array($user->role, $roles)) {
            return ResponseHelper::error(StatusCode::FORBIDDEN, 'Forbidden. You do not have the required role.');
        }

        return $next($request);
    }
}
