# texpr
Text based expression evaluator with types.

The goal of this module is to provide a way for non-technical people to learn a basic expression language and utilize it in their software. The module provides all the information necessary to provide a visual development experience where the user will be able to type and the system can estimate what they can enter next (auto-complete).  The parsing and type checking aspects of the module provide positional data to aid in communicating broken expressions.

## Features
- Types & values (simple fields, methods, or operations) are entirely user defined.
- Easy to understand left to right evaluation & parsing. ex: `user.createDate.hour`
- Types can have parameterized values (methods). ex: `today.addDays(2)`
- Type methods can be symbols which appear operation like: `today.minute+(today.hour*(60))>(120)`
- Expressions are case insensitive. ex: `TODAY=(today)`
- Basic generic support in parameterized values.
- Compilation utilities provide a way for the developer to convert expressions into a runnable function, SQL, etc.
