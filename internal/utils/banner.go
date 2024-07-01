package utils

import "github.com/fatih/color"

func ShowBanner() {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	println(red("============================================================="), green("==================================="))
	println(red("        *         * * * * *           *         *           *"), green("*     *  * * * *  * * * *  * * * * "))
	println(red("       * *        *         *        * *        * *       * *"), green("* *   *     *        *     *       "))
	println(red("      *   *       *         *       *   *       *   *   *   *"), green("*  *  *     *        *     * * * * "))
	println(red("     *     *      *         *      *     *      *     *     *"), green("*   * *     *        *     *       "))
	println(red("    * * * * *     *         *     * * * * *     *           *"), green("*     *  * * * *     *     * * * * "))
	println(red("   *         *    *         *    *         *    *           *"), green("           Version: 1.0.0          "))
	println(red("  *           *   *         *   *           *   *           *"), green("             Open Source           "))
	println(red(" *             *  * * * * *    *             *  *           *"), green(" Copyright@2021-2024 Adamnite Lab  "))
	println(red("============================================================="), green("==================================="))
}
