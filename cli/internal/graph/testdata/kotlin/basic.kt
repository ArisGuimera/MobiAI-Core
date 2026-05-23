package com.mobiai.demo

import com.x.y.Foo
import kotlinx.coroutines.flow.Flow

class LoginViewModel(private val repo: AuthRepository) {
    fun login(email: String, password: String): Flow<Result> {
        return repo.authenticate(email, password)
    }

    fun logout() {
        repo.clear()
    }
}

object Config {
    val baseUrl = "https://example.com"
}

interface AuthRepository {
    suspend fun authenticate(email: String, password: String): User?
    fun clear()
}

data class User(val id: String, val email: String)

sealed class Result {
    object Success : Result()
    data class Error(val message: String) : Result()
}
