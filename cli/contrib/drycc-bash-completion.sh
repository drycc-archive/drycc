#!/bin/bash
#
# Drycc autocomplete script for Bash.

_drycc_commands() {
  drycc help commands | cut -f 2 -d ' '
}

_drycc() {
  cur=${COMP_WORDS[COMP_CWORD]}
  prev=${COMP_WORDS[COMP_CWORD-1]}

  if [ $COMP_CWORD -eq 1 ]; then
    COMPREPLY=( $( compgen -W "$(_drycc_commands)" ${cur} ) )
  elif [ $COMP_CWORD -eq 2 ]; then
    case "${prev}" in
      help) COMPREPLY=( $( compgen -W "$(_drycc_commands)" ${cur} ) ) ;;
    esac
  fi
}

complete -F _drycc -o default drycc
