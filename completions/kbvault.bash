# bash completion for kbvault                              -*- shell-script -*-

__kbvault_debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE:-} ]]; then
        echo "$*" >> "${BASH_COMP_DEBUG_FILE}"
    fi
}

# Homebrew on Macs have version 1.3 of bash-completion which doesn't include
# _init_completion. This is a very minimal version of that function.
__kbvault_init_completion()
{
    COMPREPLY=()
    _get_comp_words_by_ref "$@" cur prev words cword
}

__kbvault_index_of_word()
{
    local w word=$1
    shift
    index=0
    for w in "$@"; do
        [[ $w = "$word" ]] && return
        index=$((index+1))
    done
    index=-1
}

__kbvault_contains_word()
{
    local w word=$1; shift
    for w in "$@"; do
        [[ $w = "$word" ]] && return
    done
    return 1
}

__kbvault_handle_go_custom_completion()
{
    __kbvault_debug "${FUNCNAME[0]}: cur is ${cur}, words[*] is ${words[*]}, #words[@] is ${#words[@]}"

    local shellCompDirectiveError=1
    local shellCompDirectiveNoSpace=2
    local shellCompDirectiveNoFileComp=4
    local shellCompDirectiveFilterFileExt=8
    local shellCompDirectiveFilterDirs=16

    local out requestComp lastParam lastChar comp directive args

    # Prepare the command to request completions for the program.
    # Calling ${words[0]} instead of directly kbvault allows handling aliases
    args=("${words[@]:1}")
    # Disable ActiveHelp which is not supported for bash completion v1
    requestComp="KBVAULT_ACTIVE_HELP=0 ${words[0]} __completeNoDesc ${args[*]}"

    lastParam=${words[$((${#words[@]}-1))]}
    lastChar=${lastParam:$((${#lastParam}-1)):1}
    __kbvault_debug "${FUNCNAME[0]}: lastParam ${lastParam}, lastChar ${lastChar}"

    if [ -z "${cur}" ] && [ "${lastChar}" != "=" ]; then
        # If the last parameter is complete (there is a space following it)
        # We add an extra empty parameter so we can indicate this to the go method.
        __kbvault_debug "${FUNCNAME[0]}: Adding extra empty parameter"
        requestComp="${requestComp} \"\""
    fi

    __kbvault_debug "${FUNCNAME[0]}: calling ${requestComp}"
    # Use eval to handle any environment variables and such
    out=$(eval "${requestComp}" 2>/dev/null)

    # Extract the directive integer at the very end of the output following a colon (:)
    directive=${out##*:}
    # Remove the directive
    out=${out%:*}
    if [ "${directive}" = "${out}" ]; then
        # There is not directive specified
        directive=0
    fi
    __kbvault_debug "${FUNCNAME[0]}: the completion directive is: ${directive}"
    __kbvault_debug "${FUNCNAME[0]}: the completions are: ${out}"

    if [ $((directive & shellCompDirectiveError)) -ne 0 ]; then
        # Error code.  No completion.
        __kbvault_debug "${FUNCNAME[0]}: received error from custom completion go code"
        return
    else
        if [ $((directive & shellCompDirectiveNoSpace)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __kbvault_debug "${FUNCNAME[0]}: activating no space"
                compopt -o nospace
            fi
        fi
        if [ $((directive & shellCompDirectiveNoFileComp)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __kbvault_debug "${FUNCNAME[0]}: activating no file completion"
                compopt +o default
            fi
        fi
    fi

    if [ $((directive & shellCompDirectiveFilterFileExt)) -ne 0 ]; then
        # File extension filtering
        local fullFilter filter filteringCmd
        # Do not use quotes around the $out variable or else newline
        # characters will be kept.
        for filter in ${out}; do
            fullFilter+="$filter|"
        done

        filteringCmd="_filedir $fullFilter"
        __kbvault_debug "File filtering command: $filteringCmd"
        $filteringCmd
    elif [ $((directive & shellCompDirectiveFilterDirs)) -ne 0 ]; then
        # File completion for directories only
        local subdir
        # Use printf to strip any trailing newline
        subdir=$(printf "%s" "${out}")
        if [ -n "$subdir" ]; then
            __kbvault_debug "Listing directories in $subdir"
            __kbvault_handle_subdirs_in_dir_flag "$subdir"
        else
            __kbvault_debug "Listing directories in ."
            _filedir -d
        fi
    else
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${out}" -- "$cur")
    fi
}

__kbvault_handle_reply()
{
    __kbvault_debug "${FUNCNAME[0]}"
    local comp
    case $cur in
        -*)
            if [[ $(type -t compopt) = "builtin" ]]; then
                compopt -o nospace
            fi
            local allflags
            if [ ${#must_have_one_flag[@]} -ne 0 ]; then
                allflags=("${must_have_one_flag[@]}")
            else
                allflags=("${flags[*]} ${two_word_flags[*]}")
            fi
            while IFS='' read -r comp; do
                COMPREPLY+=("$comp")
            done < <(compgen -W "${allflags[*]}" -- "$cur")
            if [[ $(type -t compopt) = "builtin" ]]; then
                [[ "${COMPREPLY[0]}" == *= ]] || compopt +o nospace
            fi

            # complete after --flag=abc
            if [[ $cur == *=* ]]; then
                if [[ $(type -t compopt) = "builtin" ]]; then
                    compopt +o nospace
                fi

                local index flag
                flag="${cur%=*}"
                __kbvault_index_of_word "${flag}" "${flags_with_completion[@]}"
                COMPREPLY=()
                if [[ ${index} -ge 0 ]]; then
                    PREFIX=""
                    cur="${cur#*=}"
                    ${flags_completion[${index}]}
                    if [ -n "${ZSH_VERSION:-}" ]; then
                        # zsh completion needs --flag= prefix
                        eval "COMPREPLY=( \"\${COMPREPLY[@]/#/${flag}=}\" )"
                    fi
                fi
            fi

            if [[ -z "${flag_parsing_disabled}" ]]; then
                # If flag parsing is enabled, we have completed the flags and can return.
                # If flag parsing is disabled, we may not know all (or any) of the flags, so we fallthrough
                # to possibly call handle_go_custom_completion.
                return 0;
            fi
            ;;
    esac

    # check if we are handling a flag with special work handling
    local index
    __kbvault_index_of_word "${prev}" "${flags_with_completion[@]}"
    if [[ ${index} -ge 0 ]]; then
        ${flags_completion[${index}]}
        return
    fi

    # we are parsing a flag and don't have a special handler, no completion
    if [[ ${cur} != "${words[cword]}" ]]; then
        return
    fi

    local completions
    completions=("${commands[@]}")
    if [[ ${#must_have_one_noun[@]} -ne 0 ]]; then
        completions+=("${must_have_one_noun[@]}")
    elif [[ -n "${has_completion_function}" ]]; then
        # if a go completion function is provided, defer to that function
        __kbvault_handle_go_custom_completion
    fi
    if [[ ${#must_have_one_flag[@]} -ne 0 ]]; then
        completions+=("${must_have_one_flag[@]}")
    fi
    while IFS='' read -r comp; do
        COMPREPLY+=("$comp")
    done < <(compgen -W "${completions[*]}" -- "$cur")

    if [[ ${#COMPREPLY[@]} -eq 0 && ${#noun_aliases[@]} -gt 0 && ${#must_have_one_noun[@]} -ne 0 ]]; then
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${noun_aliases[*]}" -- "$cur")
    fi

    if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
        if declare -F __kbvault_custom_func >/dev/null; then
            # try command name qualified custom func
            __kbvault_custom_func
        else
            # otherwise fall back to unqualified for compatibility
            declare -F __custom_func >/dev/null && __custom_func
        fi
    fi

    # available in bash-completion >= 2, not always present on macOS
    if declare -F __ltrim_colon_completions >/dev/null; then
        __ltrim_colon_completions "$cur"
    fi

    # If there is only 1 completion and it is a flag with an = it will be completed
    # but we don't want a space after the =
    if [[ "${#COMPREPLY[@]}" -eq "1" ]] && [[ $(type -t compopt) = "builtin" ]] && [[ "${COMPREPLY[0]}" == --*= ]]; then
       compopt -o nospace
    fi
}

# The arguments should be in the form "ext1|ext2|extn"
__kbvault_handle_filename_extension_flag()
{
    local ext="$1"
    _filedir "@(${ext})"
}

__kbvault_handle_subdirs_in_dir_flag()
{
    local dir="$1"
    pushd "${dir}" >/dev/null 2>&1 && _filedir -d && popd >/dev/null 2>&1 || return
}

__kbvault_handle_flag()
{
    __kbvault_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    # if a command required a flag, and we found it, unset must_have_one_flag()
    local flagname=${words[c]}
    local flagvalue=""
    # if the word contained an =
    if [[ ${words[c]} == *"="* ]]; then
        flagvalue=${flagname#*=} # take in as flagvalue after the =
        flagname=${flagname%=*} # strip everything after the =
        flagname="${flagname}=" # but put the = back
    fi
    __kbvault_debug "${FUNCNAME[0]}: looking for ${flagname}"
    if __kbvault_contains_word "${flagname}" "${must_have_one_flag[@]}"; then
        must_have_one_flag=()
    fi

    # if you set a flag which only applies to this command, don't show subcommands
    if __kbvault_contains_word "${flagname}" "${local_nonpersistent_flags[@]}"; then
      commands=()
    fi

    # keep flag value with flagname as flaghash
    # flaghash variable is an associative array which is only supported in bash > 3.
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        if [ -n "${flagvalue}" ] ; then
            flaghash[${flagname}]=${flagvalue}
        elif [ -n "${words[ $((c+1)) ]}" ] ; then
            flaghash[${flagname}]=${words[ $((c+1)) ]}
        else
            flaghash[${flagname}]="true" # pad "true" for bool flag
        fi
    fi

    # skip the argument to a two word flag
    if [[ ${words[c]} != *"="* ]] && __kbvault_contains_word "${words[c]}" "${two_word_flags[@]}"; then
        __kbvault_debug "${FUNCNAME[0]}: found a flag ${words[c]}, skip the next argument"
        c=$((c+1))
        # if we are looking for a flags value, don't show commands
        if [[ $c -eq $cword ]]; then
            commands=()
        fi
    fi

    c=$((c+1))

}

__kbvault_handle_noun()
{
    __kbvault_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    if __kbvault_contains_word "${words[c]}" "${must_have_one_noun[@]}"; then
        must_have_one_noun=()
    elif __kbvault_contains_word "${words[c]}" "${noun_aliases[@]}"; then
        must_have_one_noun=()
    fi

    nouns+=("${words[c]}")
    c=$((c+1))
}

__kbvault_handle_command()
{
    __kbvault_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    local next_command
    if [[ -n ${last_command} ]]; then
        next_command="_${last_command}_${words[c]//:/__}"
    else
        if [[ $c -eq 0 ]]; then
            next_command="_kbvault_root_command"
        else
            next_command="_${words[c]//:/__}"
        fi
    fi
    c=$((c+1))
    __kbvault_debug "${FUNCNAME[0]}: looking for ${next_command}"
    declare -F "$next_command" >/dev/null && $next_command
}

__kbvault_handle_word()
{
    if [[ $c -ge $cword ]]; then
        __kbvault_handle_reply
        return
    fi
    __kbvault_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"
    if [[ "${words[c]}" == -* ]]; then
        __kbvault_handle_flag
    elif __kbvault_contains_word "${words[c]}" "${commands[@]}"; then
        __kbvault_handle_command
    elif [[ $c -eq 0 ]]; then
        __kbvault_handle_command
    elif __kbvault_contains_word "${words[c]}" "${command_aliases[@]}"; then
        # aliashash variable is an associative array which is only supported in bash > 3.
        if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
            words[c]=${aliashash[${words[c]}]}
            __kbvault_handle_command
        else
            __kbvault_handle_noun
        fi
    else
        __kbvault_handle_noun
    fi
    __kbvault_handle_word
}

_kbvault_completion()
{
    last_command="kbvault_completion"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--help")
    flags+=("-h")
    local_nonpersistent_flags+=("--help")
    local_nonpersistent_flags+=("-h")
    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    must_have_one_noun+=("bash")
    must_have_one_noun+=("fish")
    must_have_one_noun+=("powershell")
    must_have_one_noun+=("zsh")
    noun_aliases=()
}

_kbvault_config_path()
{
    last_command="kbvault_config_path"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_config_set()
{
    last_command="kbvault_config_set"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_config_show()
{
    last_command="kbvault_config_show"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    local_nonpersistent_flags+=("-f")
    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_config_validate()
{
    last_command="kbvault_config_validate"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_config()
{
    last_command="kbvault_config"

    command_aliases=()

    commands=()
    commands+=("path")
    commands+=("set")
    commands+=("show")
    commands+=("validate")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_configure()
{
    last_command="kbvault_configure"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--profile=")
    two_word_flags+=("--profile")
    local_nonpersistent_flags+=("--profile")
    local_nonpersistent_flags+=("--profile=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_delete()
{
    last_command="kbvault_delete"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")
    local_nonpersistent_flags+=("-f")
    flags+=("--interactive")
    flags+=("-i")
    local_nonpersistent_flags+=("--interactive")
    local_nonpersistent_flags+=("-i")
    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_edit()
{
    last_command="kbvault_edit"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--create")
    local_nonpersistent_flags+=("--create")
    flags+=("--editor=")
    two_word_flags+=("--editor")
    local_nonpersistent_flags+=("--editor")
    local_nonpersistent_flags+=("--editor=")
    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_help()
{
    last_command="kbvault_help"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    has_completion_function=1
    noun_aliases=()
}

_kbvault_init()
{
    last_command="kbvault_init"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")
    local_nonpersistent_flags+=("-f")
    flags+=("--name=")
    two_word_flags+=("--name")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    local_nonpersistent_flags+=("-n")
    flags+=("--path=")
    two_word_flags+=("--path")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--path")
    local_nonpersistent_flags+=("--path=")
    local_nonpersistent_flags+=("-p")
    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_list()
{
    last_command="kbvault_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    local_nonpersistent_flags+=("-f")
    flags+=("--limit=")
    two_word_flags+=("--limit")
    two_word_flags+=("-l")
    local_nonpersistent_flags+=("--limit")
    local_nonpersistent_flags+=("--limit=")
    local_nonpersistent_flags+=("-l")
    flags+=("--paths")
    flags+=("-p")
    local_nonpersistent_flags+=("--paths")
    local_nonpersistent_flags+=("-p")
    flags+=("--reverse")
    flags+=("-r")
    local_nonpersistent_flags+=("--reverse")
    local_nonpersistent_flags+=("-r")
    flags+=("--sort=")
    two_word_flags+=("--sort")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--sort")
    local_nonpersistent_flags+=("--sort=")
    local_nonpersistent_flags+=("-s")
    flags+=("--tags=")
    two_word_flags+=("--tags")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--tags")
    local_nonpersistent_flags+=("--tags=")
    local_nonpersistent_flags+=("-t")
    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_new()
{
    last_command="kbvault_new"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--open")
    flags+=("-o")
    local_nonpersistent_flags+=("--open")
    local_nonpersistent_flags+=("-o")
    flags+=("--tags=")
    two_word_flags+=("--tags")
    local_nonpersistent_flags+=("--tags")
    local_nonpersistent_flags+=("--tags=")
    flags+=("--template=")
    two_word_flags+=("--template")
    local_nonpersistent_flags+=("--template")
    local_nonpersistent_flags+=("--template=")
    flags+=("--title=")
    two_word_flags+=("--title")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--title")
    local_nonpersistent_flags+=("--title=")
    local_nonpersistent_flags+=("-t")
    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_profile_copy()
{
    last_command="kbvault_profile_copy"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_profile_create()
{
    last_command="kbvault_profile_create"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--description=")
    two_word_flags+=("--description")
    local_nonpersistent_flags+=("--description")
    local_nonpersistent_flags+=("--description=")
    flags+=("--local-path=")
    two_word_flags+=("--local-path")
    local_nonpersistent_flags+=("--local-path")
    local_nonpersistent_flags+=("--local-path=")
    flags+=("--s3-bucket=")
    two_word_flags+=("--s3-bucket")
    local_nonpersistent_flags+=("--s3-bucket")
    local_nonpersistent_flags+=("--s3-bucket=")
    flags+=("--s3-endpoint=")
    two_word_flags+=("--s3-endpoint")
    local_nonpersistent_flags+=("--s3-endpoint")
    local_nonpersistent_flags+=("--s3-endpoint=")
    flags+=("--s3-region=")
    two_word_flags+=("--s3-region")
    local_nonpersistent_flags+=("--s3-region")
    local_nonpersistent_flags+=("--s3-region=")
    flags+=("--storage-type=")
    two_word_flags+=("--storage-type")
    local_nonpersistent_flags+=("--storage-type")
    local_nonpersistent_flags+=("--storage-type=")
    flags+=("--vault-name=")
    two_word_flags+=("--vault-name")
    local_nonpersistent_flags+=("--vault-name")
    local_nonpersistent_flags+=("--vault-name=")
    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_profile_delete()
{
    last_command="kbvault_profile_delete"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")
    local_nonpersistent_flags+=("-f")
    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_profile_get()
{
    last_command="kbvault_profile_get"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_profile_list()
{
    last_command="kbvault_profile_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    local_nonpersistent_flags+=("-o")
    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_profile_set()
{
    last_command="kbvault_profile_set"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_profile_show()
{
    last_command="kbvault_profile_show"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    local_nonpersistent_flags+=("-o")
    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_profile_switch()
{
    last_command="kbvault_profile_switch"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_profile()
{
    last_command="kbvault_profile"

    command_aliases=()

    commands=()
    commands+=("copy")
    commands+=("create")
    commands+=("delete")
    commands+=("get")
    commands+=("list")
    commands+=("set")
    commands+=("show")
    commands+=("switch")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_search()
{
    last_command="kbvault_search"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--after=")
    two_word_flags+=("--after")
    local_nonpersistent_flags+=("--after")
    local_nonpersistent_flags+=("--after=")
    flags+=("--before=")
    two_word_flags+=("--before")
    local_nonpersistent_flags+=("--before")
    local_nonpersistent_flags+=("--before=")
    flags+=("--build-index")
    local_nonpersistent_flags+=("--build-index")
    flags+=("--desc")
    local_nonpersistent_flags+=("--desc")
    flags+=("--detailed")
    local_nonpersistent_flags+=("--detailed")
    flags+=("--field=")
    two_word_flags+=("--field")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--field")
    local_nonpersistent_flags+=("--field=")
    local_nonpersistent_flags+=("-f")
    flags+=("--json")
    local_nonpersistent_flags+=("--json")
    flags+=("--limit=")
    two_word_flags+=("--limit")
    local_nonpersistent_flags+=("--limit")
    local_nonpersistent_flags+=("--limit=")
    flags+=("--offset=")
    two_word_flags+=("--offset")
    local_nonpersistent_flags+=("--offset")
    local_nonpersistent_flags+=("--offset=")
    flags+=("--sort=")
    two_word_flags+=("--sort")
    local_nonpersistent_flags+=("--sort")
    local_nonpersistent_flags+=("--sort=")
    flags+=("--tag=")
    two_word_flags+=("--tag")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--tag")
    local_nonpersistent_flags+=("--tag=")
    local_nonpersistent_flags+=("-t")
    flags+=("--type=")
    two_word_flags+=("--type")
    local_nonpersistent_flags+=("--type")
    local_nonpersistent_flags+=("--type=")
    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_show()
{
    last_command="kbvault_show"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--content")
    flags+=("-c")
    local_nonpersistent_flags+=("--content")
    local_nonpersistent_flags+=("-c")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    local_nonpersistent_flags+=("-f")
    flags+=("--metadata")
    flags+=("-m")
    local_nonpersistent_flags+=("--metadata")
    local_nonpersistent_flags+=("-m")
    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kbvault_root_command()
{
    last_command="kbvault"

    command_aliases=()

    commands=()
    commands+=("completion")
    commands+=("config")
    commands+=("configure")
    commands+=("delete")
    commands+=("edit")
    commands+=("help")
    commands+=("init")
    commands+=("list")
    commands+=("new")
    commands+=("profile")
    commands+=("search")
    commands+=("show")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--profile=")
    two_word_flags+=("--profile")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

__start_kbvault()
{
    local cur prev words cword split
    declare -A flaghash 2>/dev/null || :
    declare -A aliashash 2>/dev/null || :
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion -s || return
    else
        __kbvault_init_completion -n "=" || return
    fi

    local c=0
    local flag_parsing_disabled=
    local flags=()
    local two_word_flags=()
    local local_nonpersistent_flags=()
    local flags_with_completion=()
    local flags_completion=()
    local commands=("kbvault")
    local command_aliases=()
    local must_have_one_flag=()
    local must_have_one_noun=()
    local has_completion_function=""
    local last_command=""
    local nouns=()
    local noun_aliases=()

    __kbvault_handle_word
}

if [[ $(type -t compopt) = "builtin" ]]; then
    complete -o default -F __start_kbvault kbvault
else
    complete -o default -o nospace -F __start_kbvault kbvault
fi

# ex: ts=4 sw=4 et filetype=sh
