import Foundation
import Combine
import SwiftUI

class LoginViewModel: ObservableObject {
    @Published private(set) var state: LoginState = .idle
    private let repo: AuthRepository

    init(repo: AuthRepository) {
        self.repo = repo
    }

    func login(email: String, password: String) async throws -> User {
        return try await repo.authenticate(email: email, password: password)
    }

    func logout() {
        repo.clear()
    }
}

struct User: Identifiable, Codable {
    let id: String
    let email: String
}

protocol AuthRepository {
    func authenticate(email: String, password: String) async throws -> User
    func clear()
}

enum AuthError: Error {
    case invalidCredentials
    case network(underlying: Error)
}

actor SessionStore {
    private var current: User?

    func set(_ user: User) {
        current = user
    }
}

extension User {
    func displayName() -> String { email }
}
