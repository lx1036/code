//
// Created by 刘祥 on 7/1/22.
//

#ifndef BPF_TAILCALL_H
#define BPF_TAILCALL_H


#define __eval(x, ...) x ## __VA_ARGS__


#define invoke_tailcall_if(COND, NAME, FUNC)  \
	__eval(__invoke_tailcall_if_, COND)(NAME, FUNC)

#endif //BPF_TAILCALL_H
