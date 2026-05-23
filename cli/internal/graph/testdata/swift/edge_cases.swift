import Foundation

// class NotARealClass — should be ignored (line comment)

/*
class AlsoNotReal
func alsoIgnoredFn() {}
*/

class Outer {
    /* nested block /* not really */ comment */
    func transform<T>(input: String) -> T {
        let literal = "class FakeClass { func fakeFn() {} }"
        let multi = """
            class TripleQuotedFake
            func anotherFake()
            """
        return input as! T
    }

    class Inner {
        func innerMethod() {}
    }
}

extension String {
    func extensionFn() -> Int { count }
}

public final class FinalClass {
    public static func staticFn() {}
    nonisolated func isolatedFn() {}
}

@MainActor
public func mainActorFn() {}
