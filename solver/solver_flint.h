#ifndef __SOLVER_FLINT_HPP__
#define __SOLVER_FLINT_HPP__

#ifdef __cplusplus
extern "C"
{
#endif

  char ** solve(int n, char **const out_messages, const char *prime, const char **const sums);
  void *allocArgv(int argc);
#ifdef __cplusplus
}
#endif

#endif