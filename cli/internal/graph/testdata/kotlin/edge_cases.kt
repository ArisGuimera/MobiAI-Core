package com.mobiai.demo

import com.x.Y

// class NotARealClass — should be ignored (line comment)

/*
class AlsoNotReal
fun alsoIgnoredFn() {}
*/

class Outer {
    /* nested block /* not really nested */ comment */
    private inline fun <T> transform(input: String): T {
        val literal = "class FakeClass { fun fakeFn() {} }"
        val multi = """
            class TripleQuotedFake
            fun anotherFake()
        """.trimIndent()
        return input as T
    }

    inner class Inner {
        fun innerMethod() {}
    }
}

fun String.extensionFn(): Int = length

suspend fun coroutineFn() {}

@Composable
public open fun ComposableScreen() {}
