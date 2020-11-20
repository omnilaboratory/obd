package omnicore

import (
	"strconv"
	"strings"
)

/*
 * omnicore/parse_string.cpp
 * the golang code is a straight forward transalaton of cpp source code.
 */
func StrToInt64(str string, divisible bool) int64 {
	// copy original, so it remains unchanged
	var strAmount string

	strAmount = str
	var nAmount int64
	nAmount = 0

	// check for a negative (minus sign) and invalidate if present
	if strings.Contains(strAmount, "-") {
		return 0
	}

	// convert the string into a usable int64
	if divisible {
		// check for existence of decimal point
		pos := strings.Index(strAmount, ".")
		if pos == -1 {
			// no decimal point but divisible so pad 8 zeros on right
			//strAmount += "00000000";
			pad_eight_zero := "00000000"
			strAmount = strings.Join([]string{strAmount, pad_eight_zero}, "")

		} else {
			// check for existence of second decimal point, if so invalidate amount
			//size_t posSecond = strAmount.find(".", pos + 1);
			//if (posSecond != std::string::npos) return 0;
			posSecond := strings.LastIndex(strAmount, ".")
			if posSecond != pos {
				return 0
			}

			if (len(strAmount) - pos) < 9 {
				// there are decimals either exact or not enough, pad as needed
				//std::string strRightOfDecimal = strAmount.substr(pos + 1);
				//unsigned int zerosToPad = 8 - strRightOfDecimal.size();

				strRightOfDecimal := strAmount[pos+1 : len(strAmount)]
				zerosToPad := 8 - len(strRightOfDecimal)
				//fmt.Println(strRightOfDecimal)

				// do we need to pad?
				if zerosToPad > 0 {
					for it := 0; it != zerosToPad; it++ {
						strAmount += "0"
					}
				}
			} else {
				// there are too many decimals, truncate after 8
				//strAmount = strAmount.substr(0, pos + 9);
				strAmount = strAmount[0 : pos+9]
			}
			str1 := strAmount[0:pos]
			str2 := strAmount[pos+1 : len(strAmount)]
			strAmount = strings.Join([]string{str1, str2}, "")
		}
		//strAmount.erase(std::remove(strAmount.begin(), strAmount.end(), '.'), strAmount.end());
		//try {
		//    nAmount = boost::lexical_cast<int64_t>(strAmount);
		//} catch (const boost::bad_lexical_cast &e) {}

		nAmount, _ = strconv.ParseInt(strAmount, 10, 64)
		//if err != nil {
		//	fmt.Printf(strAmount)
		//	return 0
		//}

	} else {
		//   size_t pos = strAmount.find(".");
		//   std::string newStrAmount = strAmount.substr(0, pos);
		//  try {
		//       nAmount = boost::lexical_cast<int64_t>(newStrAmount);
		//   } catch (const boost::bad_lexical_cast &e) {}
		pos := strings.Index(strAmount, ".")
		if pos == -1 {
			nAmount, _ = strconv.ParseInt(strAmount, 10, 64)
		} else {
			newStrAmount := strAmount[0:pos]
			nAmount, _ = strconv.ParseInt(newStrAmount, 10, 64)
		}
	}

	return nAmount
}
