<?php

use App\Enums\ErrorCode;
use App\Enums\ResponseMessage;
use App\Helpers\ResponseHelper;
use Illuminate\Auth\Access\AuthorizationException;
use Illuminate\Auth\AuthenticationException;
use Illuminate\Database\Eloquent\ModelNotFoundException;
use Illuminate\Foundation\Application;
use Illuminate\Foundation\Configuration\Exceptions;
use Illuminate\Foundation\Configuration\Middleware;
use Illuminate\Http\Request;
use Illuminate\Validation\ValidationException;
use Symfony\Component\HttpKernel\Exception\HttpExceptionInterface;
use Symfony\Component\HttpKernel\Exception\NotFoundHttpException;

return Application::configure(basePath: dirname(__DIR__))
    ->withRouting(
        web: __DIR__.'/../routes/web.php',
        api: __DIR__.'/../routes/api.php',
        commands: __DIR__.'/../routes/console.php',
        health: '/up',
    )
    ->withMiddleware(function (Middleware $middleware) {
        $middleware->prepend(\Illuminate\Http\Middleware\HandleCors::class);

        // API only: never redirect unauthenticated requests to a login page.
        $middleware->redirectGuestsTo(fn () => null);

        $middleware->alias([
            'role' => \App\Http\Middleware\RoleMiddleware::class,
        ]);
    })
    ->withExceptions(function (Exceptions $exceptions) {
        // Render every API error in the standard { success, message, error_code, data, errors? } shape.
        $exceptions->render(function (Throwable $e, Request $request) {
            if (! $request->is('api/*') && ! $request->expectsJson()) {
                return null;
            }

            return match (true) {
                $e instanceof ValidationException => ResponseHelper::error($e->status, $e->getMessage(), $e->errors(), ErrorCode::VALIDATION_FAILED),
                $e instanceof AuthenticationException => ResponseHelper::error(401, 'Unauthenticated.'),
                $e instanceof AuthorizationException => ResponseHelper::error(403, $e->getMessage() ?: 'Forbidden.'),
                $e instanceof ModelNotFoundException,
                $e instanceof NotFoundHttpException => ResponseHelper::error(404, ResponseMessage::NO_ITEM),
                $e instanceof HttpExceptionInterface => ResponseHelper::error($e->getStatusCode(), $e->getMessage() ?: null),
                default => ResponseHelper::error(500, config('app.debug') ? $e->getMessage() : ResponseMessage::INTERNAL_SERVER_ERROR),
            };
        });
    })->create();
