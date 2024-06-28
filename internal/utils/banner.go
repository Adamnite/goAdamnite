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
	println(red("   *         *    *         *    *         *    *           *"), green("                                   "))
	println(red("  *           *   *         *   *           *   *           *"), green("                                   "))
	println(red(" *             *  * * * * *    *             *  *           *"), green("    Copyright@2024 Adamnite Lab    "))
	println(red("============================================================="), green("==================================="))
}
