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

func ShowBootNodeBanner() {
	yellow := color.New(color.FgHiYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	println(green("======================================================================"))
	println(green("  * *       *        *     * * * *  *     *     *     * *    * * * *  "))
	println(green("  *   *   *   *    *   *      *     * *   *   *   *   *   *  *        "))
	println(green("  * *    *     *  *     *     *     *  *  *  *     *  *    * * * * *  "))
	println(green("  *   *   *   *    *   *      *     *   * *   *   *   *   *  *        "))
	println(green("  * *       *        *        *     *     *     *     * *    * * * *  "))
	println(yellow("                       Adamnite Version: 1.0.0                        "))
	println(yellow("                             Open Source                              "))
	println(yellow("                    Copyright@2021-2024 Adamnite Lab                  "))
	println(green("======================================================================"))
}
